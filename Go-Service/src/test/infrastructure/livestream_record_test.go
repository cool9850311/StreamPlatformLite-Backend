package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"Go-Service/src/main/application/usecase"
	chatEntity "Go-Service/src/main/domain/entity/chat"
	livestreamEntity "Go-Service/src/main/domain/entity/livestream"
	"Go-Service/src/main/infrastructure/cache"
	"Go-Service/src/main/infrastructure/config"
	"Go-Service/src/main/infrastructure/controller"
	"Go-Service/src/main/infrastructure/middleware"
	"Go-Service/src/main/infrastructure/util"

	claims "github.com/cool9850311/StreamPlatformLite-Core/pkg/claims"
	"github.com/cool9850311/StreamPlatformLite-Core/pkg/role"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ---------------------------------------------------------------------------
// Test bootstrap
// ---------------------------------------------------------------------------

const recordTestSecret = "m6zr8Z1NL3ctUi2lcF8QEtZxI"

func init() {
	config.AppConfig.JWT.SecretKey = recordTestSecret
	gin.SetMode(gin.TestMode)
}

// ---------------------------------------------------------------------------
// Local stubs.
//
// Only the dependencies that would otherwise need real infrastructure (db,
// redis, rtmp server, ffmpeg binary) are stubbed. The file cache is the REAL
// cache.NewFileCache() so the tests exercise real filesystem behaviour.
// ---------------------------------------------------------------------------

type recordTestLogger struct{}

func (l *recordTestLogger) Panic(ctx context.Context, msg string) {}
func (l *recordTestLogger) Fatal(ctx context.Context, msg string) {}
func (l *recordTestLogger) Error(ctx context.Context, msg string) {}
func (l *recordTestLogger) Warn(ctx context.Context, msg string)  {}
func (l *recordTestLogger) Info(ctx context.Context, msg string)  {}
func (l *recordTestLogger) Debug(ctx context.Context, msg string) {}
func (l *recordTestLogger) Trace(ctx context.Context, msg string) {}

// recordTestDefaultTitle is the livestream Title seen by every test that does
// not care about the display name. Since the storage/display decoupling the
// Title - not the on-disk name - is what drives Content-Disposition and the
// status endpoint's "filename" field, so it has to be configurable per test.
const recordTestDefaultTitle = "test-record"

type recordTestRepo struct {
	title string
}

func (r *recordTestRepo) GetByID(id string) (*livestreamEntity.Livestream, error) {
	return &livestreamEntity.Livestream{UUID: id, Title: r.title}, nil
}
func (r *recordTestRepo) GetByOwnerID(ownerID string) (*livestreamEntity.Livestream, error) {
	return &livestreamEntity.Livestream{UUID: ownerID, Title: r.title}, nil
}
func (r *recordTestRepo) GetOne() (*livestreamEntity.Livestream, error) {
	return &livestreamEntity.Livestream{Title: r.title}, nil
}
func (r *recordTestRepo) Create(*livestreamEntity.Livestream) error { return nil }
func (r *recordTestRepo) Update(*livestreamEntity.Livestream) error { return nil }
func (r *recordTestRepo) Delete(string) error                       { return nil }
func (r *recordTestRepo) MuteUser(string, string, string) error     { return nil }

type recordTestStreamService struct{}

func (s *recordTestStreamService) OpenStream(name, uuid, apiKey string, isRecord bool) error {
	return nil
}
func (s *recordTestStreamService) CloseStream(uuid string) error      { return nil }
func (s *recordTestStreamService) StartService() error                { return nil }
func (s *recordTestStreamService) RunLoop() error                     { return nil }
func (s *recordTestStreamService) IsLiveStreamExist(uuid string) bool { return true }

type recordTestViewerCount struct{}

func (v *recordTestViewerCount) GetViewerCount(livestreamUUID string) (int, error) { return 0, nil }
func (v *recordTestViewerCount) AddViewerCount(livestreamUUID, userID string) error {
	return nil
}
func (v *recordTestViewerCount) RemoveViewerCount(livestreamUUID string, seconds int) (int, error) {
	return 0, nil
}

type recordTestChatCache struct{}

func (c *recordTestChatCache) GetChat(livestreamUUID, index string, count int) ([]chatEntity.Chat, error) {
	return nil, nil
}
func (c *recordTestChatCache) AddChat(livestreamUUID string, ch chatEntity.Chat) error { return nil }
func (c *recordTestChatCache) DeleteChat(livestreamUUID, chatID string) error          { return nil }
func (c *recordTestChatCache) GetDeleteChatIDs(livestreamUUID string) ([]string, error) {
	return nil, nil
}
func (c *recordTestChatCache) GetChatByID(livestreamUUID, chatID string) (*chatEntity.Chat, error) {
	return nil, nil
}

// recordTestFfmpeg is a no-op stand-in for the ffmpeg library so the background
// conversion goroutine started by GetRecordStatus never shells out to ffmpeg.
type recordTestFfmpeg struct{}

func (f *recordTestFfmpeg) ConvertStreamToMp4(filePath string) error { return nil }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testRootPath resolves the project root exactly the way the controller does,
// i.e. through util.GetProjectRootPath(), which derives it from the CWD of the
// running test binary (this package's source directory).
func testRootPath(t *testing.T) string {
	t.Helper()
	rootPath, err := util.GetProjectRootPath()
	if err != nil {
		t.Fatalf("GetProjectRootPath failed: %v", err)
	}
	if _, err := os.Stat(rootPath); err != nil {
		t.Fatalf("resolved project root %q is not usable: %v", rootPath, err)
	}
	return rootPath
}

// recordContent returns 100 deterministic bytes (0x00..0x63).
func recordContent() []byte {
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}
	return data
}

// writeHLSFile creates <root>/hls/<uuid>/<name> with the given content and
// registers cleanup for everything it created.
func writeHLSFile(t *testing.T, uuidStr, name string, content []byte) string {
	t.Helper()
	root := testRootPath(t)
	hlsDir := filepath.Join(root, "hls")
	streamDir := filepath.Join(hlsDir, uuidStr)

	if err := os.MkdirAll(streamDir, 0o755); err != nil {
		t.Fatalf("failed to create %s: %v", streamDir, err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(streamDir)
		// Only removes hls/ when it is empty, i.e. when the test created it.
		_ = os.Remove(hlsDir)
	})

	fullPath := filepath.Join(streamDir, name)
	if err := os.WriteFile(fullPath, content, 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", fullPath, err)
	}
	return fullPath
}

// makeEmptyStreamDir creates <root>/hls/<uuid>/ with nothing in it.
func makeEmptyStreamDir(t *testing.T, uuidStr string) string {
	t.Helper()
	root := testRootPath(t)
	hlsDir := filepath.Join(root, "hls")
	streamDir := filepath.Join(hlsDir, uuidStr)
	if err := os.MkdirAll(streamDir, 0o755); err != nil {
		t.Fatalf("failed to create %s: %v", streamDir, err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(streamDir)
		_ = os.Remove(hlsDir)
	})
	return streamDir
}

func adminToken(t *testing.T) string {
	t.Helper()
	cl := &claims.Claims{
		UserID:   "admin-user",
		UserName: "Admin",
		Role:     role.Admin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, cl).SignedString([]byte(recordTestSecret))
	if err != nil {
		t.Fatalf("failed to sign test JWT: %v", err)
	}
	return token
}

func newRecordUsecase(title string) *usecase.LivestreamUsecase {
	return usecase.NewLivestreamUsecase(
		&recordTestRepo{title: title},
		&recordTestLogger{},
		config.AppConfig,
		&recordTestStreamService{},
		&recordTestViewerCount{},
		&recordTestChatCache{},
		cache.NewFileCache(), // real filesystem-backed cache
		&recordTestFfmpeg{},
	)
}

// newRecordRouter builds a gin engine with the two record endpoints wired the
// same way router.go wires them: real JWTAuthMiddleware + real controller.
func newRecordRouter(title string) *gin.Engine {
	log := &recordTestLogger{}
	ctrl := controller.NewLivestreamController(log, newRecordUsecase(title), nil)

	r := gin.New()
	group := r.Group("/livestream")
	group.GET("/record/:uuid", middleware.JWTAuthMiddleware(log), ctrl.GetRecord)
	group.GET("/record/:uuid/status", middleware.JWTAuthMiddleware(log), ctrl.GetRecordStatus)
	return r
}

// doRecordRequest issues a GET against the record router, optionally with an
// admin cookie and a Range header, using the default livestream Title.
func doRecordRequest(t *testing.T, path string, withAuth bool, rangeHeader string) *httptest.ResponseRecorder {
	t.Helper()
	return doRecordRequestWithTitle(t, path, withAuth, rangeHeader, recordTestDefaultTitle)
}

// doRecordRequestWithTitle is doRecordRequest with the livestream Title under
// test, which is what the display name is now derived from.
func doRecordRequestWithTitle(t *testing.T, path string, withAuth bool, rangeHeader, title string) *httptest.ResponseRecorder {
	t.Helper()
	r := newRecordRouter(title)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if withAuth {
		req.AddCookie(&http.Cookie{Name: "token", Value: adminToken(t)})
	}
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeJSONBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode JSON body %q: %v", w.Body.String(), err)
	}
	return body
}

// ---------------------------------------------------------------------------
// 1. GET record without Range
// ---------------------------------------------------------------------------

func TestGetRecord_SetsAcceptRanges(t *testing.T) {
	const streamUUID = "11111111-1111-4111-8111-111111111111"
	content := recordContent()
	writeHLSFile(t, streamUUID, "record.mp4", content)

	w := doRecordRequest(t, "/livestream/record/"+streamUUID, true, "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Accept-Ranges"); got != "bytes" {
		t.Fatalf("expected Accept-Ranges 'bytes', got %q", got)
	}
	if got := w.Header().Get("Content-Length"); got != "100" {
		t.Fatalf("expected Content-Length '100', got %q", got)
	}
	cd := w.Header().Get("Content-Disposition")
	if !regexp.MustCompile(`filename\*=UTF-8''`).MatchString(cd) {
		t.Fatalf("expected Content-Disposition to contain RFC 5987 form, got %q", cd)
	}
	if got := w.Body.Bytes(); string(got) != string(content) {
		t.Fatalf("body mismatch: got %d bytes, want %d identical bytes", len(got), len(content))
	}
}

// ---------------------------------------------------------------------------
// 2. Byte range
// ---------------------------------------------------------------------------

func TestGetRecord_Range_Returns206(t *testing.T) {
	const streamUUID = "22222222-2222-4222-8222-222222222222"
	content := recordContent()
	writeHLSFile(t, streamUUID, "record.mp4", content)

	w := doRecordRequest(t, "/livestream/record/"+streamUUID, true, "bytes=10-19")

	if w.Code != http.StatusPartialContent {
		t.Fatalf("expected 206, got %d (body: %s)", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Content-Range"); got != "bytes 10-19/100" {
		t.Fatalf("expected Content-Range 'bytes 10-19/100', got %q", got)
	}
	if got := w.Header().Get("Content-Length"); got != "10" {
		t.Fatalf("expected Content-Length '10', got %q", got)
	}
	if got, want := w.Body.String(), string(content[10:20]); got != want {
		t.Fatalf("body mismatch: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// 3. Suffix range
// ---------------------------------------------------------------------------

func TestGetRecord_Range_Suffix(t *testing.T) {
	const streamUUID = "33333333-3333-4333-8333-333333333333"
	content := recordContent()
	writeHLSFile(t, streamUUID, "record.mp4", content)

	w := doRecordRequest(t, "/livestream/record/"+streamUUID, true, "bytes=-5")

	if w.Code != http.StatusPartialContent {
		t.Fatalf("expected 206, got %d (body: %s)", w.Code, w.Body.String())
	}
	if got, want := w.Body.String(), string(content[95:]); got != want {
		t.Fatalf("body mismatch: got %q, want %q", got, want)
	}
	if got := w.Header().Get("Content-Range"); got != "bytes 95-99/100" {
		t.Fatalf("expected Content-Range 'bytes 95-99/100', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// 4. Unsatisfiable range
// ---------------------------------------------------------------------------

func TestGetRecord_Range_Unsatisfiable(t *testing.T) {
	const streamUUID = "44444444-4444-4444-8444-444444444444"
	writeHLSFile(t, streamUUID, "record.mp4", recordContent())

	w := doRecordRequest(t, "/livestream/record/"+streamUUID, true, "bytes=500-600")

	if w.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("expected 416, got %d (body: %s)", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Content-Range"); got != "bytes */100" {
		t.Fatalf("expected Content-Range 'bytes */100', got %q", got)
	}
}

// ---------------------------------------------------------------------------
// 5. UTF-8 filename
// ---------------------------------------------------------------------------

// The display name is now derived from the livestream Title, so the multi-byte
// value under test moved from the on-disk name to the Title. The purpose is
// unchanged: prove a non-ASCII name survives RFC 5987 encoding end-to-end,
// including the frontend's own extraction regex.
func TestGetRecord_UTF8Filename(t *testing.T) {
	const streamUUID = "55555555-5555-4555-8555-555555555555"
	const title = "直播錄影"
	writeHLSFile(t, streamUUID, util.RecordingFileName, recordContent())

	// Sanity: the sanitised Title really is multi-byte, otherwise this test
	// would silently stop exercising percent-encoding of UTF-8 sequences.
	originalName := util.SanitizeDownloadFilename(title)
	if originalName != "直播錄影.mp4" {
		t.Fatalf("unexpected sanitised display name %q", originalName)
	}

	w := doRecordRequestWithTitle(t, "/livestream/record/"+streamUUID, true, "", title)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	cd := w.Header().Get("Content-Disposition")
	expected := "attachment; filename*=UTF-8''" + util.EncodeRFC5987(originalName)
	if cd != expected {
		t.Fatalf("Content-Disposition mismatch:\n got: %q\nwant: %q", cd, expected)
	}

	// The exact regex the frontend uses to recover the suggested filename.
	matches := regexp.MustCompile(`filename\*=UTF-8''([^;]+)`).FindStringSubmatch(cd)
	if len(matches) != 2 {
		t.Fatalf("frontend regex did not match Content-Disposition %q", cd)
	}
	encoded := matches[1]

	decodedQuery, err := url.QueryUnescape(encoded)
	if err != nil {
		t.Fatalf("QueryUnescape(%q) failed: %v", encoded, err)
	}
	if decodedQuery != originalName {
		t.Fatalf("QueryUnescape round-trip mismatch: got %q, want %q", decodedQuery, originalName)
	}

	decodedPath, err := url.PathUnescape(encoded)
	if err != nil {
		t.Fatalf("PathUnescape(%q) failed: %v", encoded, err)
	}
	if decodedPath != originalName {
		t.Fatalf("PathUnescape round-trip mismatch: got %q, want %q", decodedPath, originalName)
	}
}

// ---------------------------------------------------------------------------
// 6. Status: ready
// ---------------------------------------------------------------------------

func TestGetRecordStatus_Ready_200(t *testing.T) {
	const streamUUID = "66666666-6666-4666-8666-666666666666"
	const filename = "record.mp4"
	content := recordContent()
	fullPath := writeHLSFile(t, streamUUID, filename, content)

	info, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("stat %s failed: %v", fullPath, err)
	}

	w := doRecordRequest(t, "/livestream/record/"+streamUUID+"/status", true, "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
	body := decodeJSONBody(t, w)
	if body["state"] != usecase.RecordStateReady {
		t.Fatalf("expected state %q, got %v", usecase.RecordStateReady, body["state"])
	}
	// "filename" is the display name, which is derived from the livestream
	// Title and is deliberately independent of the on-disk name seeded above.
	wantFilename := util.SanitizeDownloadFilename(recordTestDefaultTitle)
	if body["filename"] != wantFilename {
		t.Fatalf("expected filename %q, got %v", wantFilename, body["filename"])
	}
	size, ok := body["size"].(float64)
	if !ok {
		t.Fatalf("expected numeric size, got %T (%v)", body["size"], body["size"])
	}
	if int64(size) != info.Size() {
		t.Fatalf("expected size %d, got %d", info.Size(), int64(size))
	}
	if int64(size) != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), int64(size))
	}
}

// ---------------------------------------------------------------------------
// 7. Status: converting
// ---------------------------------------------------------------------------

func TestGetRecordStatus_Converting_202(t *testing.T) {
	const streamUUID = "77777777-7777-4777-8777-777777777777"
	// A record playlist but no mp4 -> a conversion is kicked off in the
	// background against the no-op ffmpeg stub.
	writeHLSFile(t, streamUUID, "record.m3u8", []byte("#EXTM3U\n#EXT-X-ENDLIST\n"))

	w := doRecordRequest(t, "/livestream/record/"+streamUUID+"/status", true, "")

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d (body: %s)", w.Code, w.Body.String())
	}
	body := decodeJSONBody(t, w)
	if body["state"] != usecase.RecordStateConverting {
		t.Fatalf("expected state %q, got %v", usecase.RecordStateConverting, body["state"])
	}
	elapsed, ok := body["elapsed_seconds"]
	if !ok {
		t.Fatalf("expected elapsed_seconds field, got %v", body)
	}
	if _, ok := elapsed.(float64); !ok {
		t.Fatalf("expected numeric elapsed_seconds, got %T (%v)", elapsed, elapsed)
	}
}

// ---------------------------------------------------------------------------
// 8. Status: unavailable
// ---------------------------------------------------------------------------

func TestGetRecordStatus_Unavailable_404(t *testing.T) {
	const streamUUID = "88888888-8888-4888-8888-888888888888"
	// Directory exists but holds neither an mp4 nor record.m3u8.
	makeEmptyStreamDir(t, streamUUID)

	w := doRecordRequest(t, "/livestream/record/"+streamUUID+"/status", true, "")

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
	body := decodeJSONBody(t, w)
	if body["state"] != usecase.RecordStateUnavailable {
		t.Fatalf("expected state %q, got %v", usecase.RecordStateUnavailable, body["state"])
	}
}

// ---------------------------------------------------------------------------
// 9. Status: no cookie
// ---------------------------------------------------------------------------

func TestGetRecordStatus_NoCookie_401(t *testing.T) {
	const streamUUID = "99999999-9999-4999-8999-999999999999"

	w := doRecordRequest(t, "/livestream/record/"+streamUUID+"/status", false, "")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// 10. Status: invalid uuid
// ---------------------------------------------------------------------------

func TestGetRecordStatus_BadUUID_400(t *testing.T) {
	w := doRecordRequest(t, "/livestream/record/not-a-uuid/status", true, "")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// 11. Route table registration
// ---------------------------------------------------------------------------

func TestRouterRegistration_NoPanic(t *testing.T) {
	log := &recordTestLogger{}

	// markHandler records which handler ran plus the extracted :uuid param.
	markHandler := func(name string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"handler": name, "uuid": c.Param("uuid")})
		}
	}

	var r *gin.Engine
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				t.Fatalf("route registration panicked: %v", rec)
			}
		}()

		r = gin.New()
		livestream := r.Group("/livestream")
		{
			livestream.GET("/:uuid/:filename", middleware.OptionalJWTAuthMiddleware(log), markHandler("GetFile"))
			livestream.GET("/one", middleware.OptionalJWTAuthMiddleware(log), markHandler("GetLivestreamOne"))
			livestream.GET("/:uuid", middleware.OptionalJWTAuthMiddleware(log), markHandler("GetLivestreamByID"))
			livestream.GET("/ping-viewer-count/:uuid", middleware.OptionalJWTAuthMiddleware(log), markHandler("PingViewerCount"))

			livestream.POST("", middleware.JWTAuthMiddleware(log), markHandler("CreateLivestream"))
			livestream.PATCH("/:uuid", middleware.JWTAuthMiddleware(log), markHandler("UpdateLivestream"))
			livestream.DELETE("/:uuid", middleware.JWTAuthMiddleware(log), markHandler("DeleteLivestream"))
			livestream.GET("/owner/:user_id", middleware.JWTAuthMiddleware(log), markHandler("GetLivestreamByOwnerId"))
			livestream.GET("/record/:uuid", middleware.JWTAuthMiddleware(log), markHandler("GetRecord"))
			livestream.GET("/record/:uuid/status", middleware.JWTAuthMiddleware(log), markHandler("GetRecordStatus"))

			chat := livestream.Group("/chat")
			{
				chat.GET("/:uuid/:index", middleware.OptionalJWTAuthMiddleware(log), markHandler("GetChat"))
				chat.GET("/delete/:uuid", middleware.OptionalJWTAuthMiddleware(log), markHandler("GetDeleteChatIDs"))
				chat.POST("", middleware.JWTAuthMiddleware(log), markHandler("AddChat"))
				chat.DELETE("/:uuid/:chat_id", middleware.JWTAuthMiddleware(log), markHandler("RemoveViewerCount"))
			}

			livestream.POST("/mute-user", middleware.JWTAuthMiddleware(log), markHandler("MuteUser"))
		}
	}()

	const streamUUID = "abcdabcd-abcd-4bcd-8bcd-abcdabcdabcd"
	token := adminToken(t)

	cases := []struct {
		name        string
		path        string
		wantHandler string
		wantUUID    string
	}{
		{"hls playlist", "/livestream/" + streamUUID + "/index.m3u8", "GetFile", streamUUID},
		{"record download", "/livestream/record/" + streamUUID, "GetRecord", streamUUID},
		{"record status", "/livestream/record/" + streamUUID + "/status", "GetRecordStatus", streamUUID},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: token})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d (body: %s)", tc.name, w.Code, w.Body.String())
		}
		body := decodeJSONBody(t, w)
		if body["handler"] != tc.wantHandler {
			t.Fatalf("%s: expected handler %q, got %v", tc.name, tc.wantHandler, body["handler"])
		}
		if body["uuid"] != tc.wantUUID {
			t.Fatalf("%s: expected uuid %q, got %v", tc.name, tc.wantUUID, body["uuid"])
		}
	}

	// Sanity: the three paths really did reach three distinct handlers.
	seen := map[string]bool{}
	for _, tc := range cases {
		if seen[tc.wantHandler] {
			t.Fatalf("handler %s expected more than once", tc.wantHandler)
		}
		seen[tc.wantHandler] = true
	}
	if len(seen) != 3 {
		t.Fatalf("expected 3 distinct handlers, got %d", len(seen))
	}
}

// ---------------------------------------------------------------------------
// 12. Display name comes from the Title, not from the fixed storage name
// ---------------------------------------------------------------------------

// D-1. Recordings are written under the compile-time constant
// util.RecordingFileName, so the on-disk name carries no user-facing meaning
// any more. This pins the decoupling: the browser must be offered the Title,
// and must never see the storage name.
func TestGetRecord_ContentDispositionFromTitle(t *testing.T) {
	const streamUUID = "aaaaaaaa-1111-4111-8111-aaaaaaaaaaaa"
	const title = "我的直播"

	writeHLSFile(t, streamUUID, util.RecordingFileName, recordContent())

	w := doRecordRequestWithTitle(t, "/livestream/record/"+streamUUID, true, "", title)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	cd := w.Header().Get("Content-Disposition")
	want := "attachment; filename*=UTF-8''" + util.EncodeRFC5987(util.SanitizeDownloadFilename(title))
	if cd != want {
		t.Fatalf("Content-Disposition mismatch:\n got: %q\nwant: %q", cd, want)
	}

	matches := regexp.MustCompile(`filename\*=UTF-8''([^;]+)`).FindStringSubmatch(cd)
	if len(matches) != 2 {
		t.Fatalf("frontend regex did not match Content-Disposition %q", cd)
	}
	decoded, err := url.QueryUnescape(matches[1])
	if err != nil {
		t.Fatalf("QueryUnescape(%q) failed: %v", matches[1], err)
	}
	if decoded != util.SanitizeDownloadFilename(title) {
		t.Fatalf("decoded filename mismatch: got %q, want %q", decoded, util.SanitizeDownloadFilename(title))
	}

	// The whole point: the storage name never leaks to the client.
	if decoded == util.RecordingFileName {
		t.Fatalf("Content-Disposition exposed the on-disk storage name %q instead of the Title", util.RecordingFileName)
	}
	if strings.Contains(cd, util.RecordingFileName) {
		t.Fatalf("Content-Disposition %q must not contain the storage name %q", cd, util.RecordingFileName)
	}
}

// ---------------------------------------------------------------------------
// 13. Hostile Title cannot break out of the response header
// ---------------------------------------------------------------------------

// D-2. A Title is admin-controlled free text and reaches the response header.
// CR/LF in it would be response splitting and a path separator would be a
// traversal attempt in any consumer that treats the value as a path, so both
// classes have to be gone from the bytes actually put on the wire.
func TestGetRecord_HostileTitleHeaderIsSafe(t *testing.T) {
	const streamUUID = "bbbbbbbb-2222-4222-8222-bbbbbbbbbbbb"
	const hostileTitle = "evil\r\nSet-Cookie: pwned=1/../x"

	writeHLSFile(t, streamUUID, util.RecordingFileName, recordContent())

	w := doRecordRequestWithTitle(t, "/livestream/record/"+streamUUID, true, "", hostileTitle)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	cd := w.Header().Get("Content-Disposition")

	// Raw bytes of the header value: no CR (0x0D) and no LF (0x0A) anywhere.
	for i, b := range []byte(cd) {
		if b == 0x0D || b == 0x0A {
			t.Fatalf("Content-Disposition contains a raw control byte 0x%02X at offset %d: %q", b, i, cd)
		}
	}
	// Also reject a separator surviving into the header value.
	if strings.ContainsAny(cd, "/\\") {
		t.Fatalf("Content-Disposition contains a path separator: %q", cd)
	}

	// Nothing was smuggled into the response as a new header.
	res := w.Result()
	defer res.Body.Close()
	if got := res.Header.Values("Set-Cookie"); len(got) != 0 {
		t.Fatalf("hostile Title injected Set-Cookie header(s): %v", got)
	}
	for name, values := range res.Header {
		for _, v := range values {
			if strings.ContainsAny(v, "\r\n") {
				t.Fatalf("header %q value contains CR/LF: %q", name, v)
			}
		}
	}

	// The client still gets the sanitised Title back, intact.
	want := util.SanitizeDownloadFilename(hostileTitle)
	if cd != "attachment; filename*=UTF-8''"+util.EncodeRFC5987(want) {
		t.Fatalf("Content-Disposition mismatch:\n got: %q\nwant: %q",
			cd, "attachment; filename*=UTF-8''"+util.EncodeRFC5987(want))
	}
	matches := regexp.MustCompile(`filename\*=UTF-8''([^;]+)`).FindStringSubmatch(cd)
	if len(matches) != 2 {
		t.Fatalf("frontend regex did not match Content-Disposition %q", cd)
	}
	decoded, err := url.QueryUnescape(matches[1])
	if err != nil {
		t.Fatalf("QueryUnescape(%q) failed: %v", matches[1], err)
	}
	if decoded != want {
		t.Fatalf("decoded filename mismatch: got %q, want %q", decoded, want)
	}
}

// ---------------------------------------------------------------------------
// 14. Legacy, Title-named recordings stay downloadable
// ---------------------------------------------------------------------------

// D-3. Recordings produced before the fixed-name change are still on disk under
// their old Title-derived names and are never migrated, so the "*.mp4" glob
// fallback has to keep serving them byte for byte.
func TestGetRecord_LegacyDiskFileStillDownloadable(t *testing.T) {
	const streamUUID = "cccccccc-3333-4333-8333-cccccccccccc"
	const legacyName = "阿豹的ASMR.mp4"

	content := recordContent()
	writeHLSFile(t, streamUUID, legacyName, content)

	// Guard the premise: only the legacy file exists, no fixed-name recording.
	streamDir := filepath.Join(testRootPath(t), "hls", streamUUID)
	if _, err := os.Stat(filepath.Join(streamDir, util.RecordingFileName)); !os.IsNotExist(err) {
		t.Fatalf("expected no %s in %s (stat err: %v)", util.RecordingFileName, streamDir, err)
	}

	w := doRecordRequestWithTitle(t, "/livestream/record/"+streamUUID, true, "", recordTestDefaultTitle)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
	if got := w.Body.Bytes(); !bytes.Equal(got, content) {
		t.Fatalf("body mismatch: got %d bytes %v, want %d bytes %v",
			len(got), got, len(content), content)
	}
	if got := w.Header().Get("Content-Length"); got != "100" {
		t.Fatalf("expected Content-Length '100', got %q", got)
	}
}
