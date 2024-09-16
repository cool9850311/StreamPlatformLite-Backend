package livestream

import (
	"Go-Service/src/main/domain/entity/errors"
	"Go-Service/src/main/domain/interface/logger"
	"context"
	"net"
	"sync"
	"strings"
	"Go-Service/src/main/infrastructure/util"
	"github.com/cool9850311/lal-StreamPlatformLite/pkg/rtmp"
	"path/filepath"
	"os"
	"github.com/cool9850311/lal-StreamPlatformLite/pkg/base"
	"github.com/cool9850311/lal-StreamPlatformLite/pkg/hls"
	"github.com/cool9850311/lal-StreamPlatformLite/pkg/remux"
)


type LivestreamService struct {
	listener net.Listener
	logger   logger.Logger
	streams  map[string]*livestream
}
type livestream struct {
	name      string
	uuid      string
	conn      net.Conn
	apiKey    string
	outputPathUUID string
}

func NewLivestreamService(logger logger.Logger) *LivestreamService {
	return &LivestreamService{logger: logger, streams: make(map[string]*livestream)}
}

func (l *LivestreamService) StartService() error {
	var err error
	l.listener, err = net.Listen("tcp", ":1935")
	if err != nil {
		return errors.ErrConnectionClosed
	}
	l.logger.Info(context.TODO(), "start rtmp server listen. addr= "+"1935")
	return nil
}
func (l *LivestreamService) RunLoop() error {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			return err
		}
		go l.handleTcpConnect(conn)
	}
}
func (l *LivestreamService) handleTcpConnect(conn net.Conn) error {
	remoteAddr := conn.RemoteAddr().String()
	l.logger.Info(context.TODO(), "accept a rtmp connection. remoteAddr="+remoteAddr)

	session := rtmp.NewServerSession(conn)
	var rtmp2Mpegts *remux.Rtmp2MpegtsRemuxer
	var once sync.Once

	task := func(stream *rtmp.Stream) error {
		switch stream.Header.MsgTypeId {
		case base.RtmpTypeIdCommandMessageAmf0:
			_ = session.DoCommandMessage(stream)
			if session.Url() == "" {
				break
			}

			stream, found := l.getLivestreamByUrl(session.Url())
			if !found {
				session.Dispose()
				l.logger.Warn(context.TODO(), "Unauthorized livestream attempt: " + session.Url())
				return nil
			}

			once.Do(func() {
				rootPath, err := util.GetProjectRootPath()
				if err != nil {
					l.logger.Error(context.TODO(), "Failed to get project root path: " + err.Error())
					return
				}
				outputPath := rootPath + "/hls/" + stream.outputPathUUID
				hlsMuxerConfig := hls.MuxerConfig{
					OutPath:            outputPath,
					FragmentDurationMs: 500,
					FragmentNum:        20,
					CleanupMode:        2,
				}
				hlsMuxer := hls.NewMuxer(stream.name, &hlsMuxerConfig, nil)
				hlsMuxer.Start()
				rtmp2Mpegts = remux.NewRtmp2MpegtsRemuxer(hlsMuxer)
				stream.conn = conn
				l.logger.Info(context.TODO(), "Started livestream: %s"+ stream.name)
			})
		case base.RtmpTypeIdWinAckSize:
			_ = session.DoWinAckSize(stream)
		case base.RtmpTypeIdSetChunkSize:
			// noop
		case base.RtmpTypeIdCommandMessageAmf3:
			_ = session.DoCommandAmf3Message(stream)
		case base.RtmpTypeIdMetadata:
			_ = session.DoDataMessageAmf0(stream)
		case base.RtmpTypeIdAck:
			_ = session.DoAck(stream)
		case base.RtmpTypeIdUserControl:
			_ = session.DoUserControl(stream)
		}

		if rtmp2Mpegts != nil {
			rtmp2Mpegts.FeedRtmpMessage(stream.ToAvMsg())
		}
		return nil
	}
	_ = session.RunLoop(task)
	session.Dispose()
	return nil
}
func (l *LivestreamService) getLivestreamByUrl(url string) (*livestream, bool) {
	for _, stream := range l.streams {
		if stream.apiKey != "" && strings.Contains(url, stream.apiKey) {
			return stream, true
		}
	}
	return nil, false
}
func (l *LivestreamService) IsLiveStreamExist(uuid string) bool {
	_, exists := l.streams[uuid]
	return exists
}


func (l *LivestreamService) OpenStream(name, uuid, apiKey string, outputPathUUID string) error {
	// Create a new livestream instance
	newStream := &livestream{
		name:      name,
		uuid:      uuid,
		apiKey:    apiKey,
		outputPathUUID: outputPathUUID,
	}

	l.streams[uuid] = newStream


	return nil
}

func (l *LivestreamService) CloseStream(uuid string) error {
	if stream, exists := l.streams[uuid]; exists {
		delete(l.streams, uuid)
		// Delete the HLS directory for the closed stream
		rootPath, err := util.GetProjectRootPath()
		if err != nil {
			l.logger.Error(context.TODO(), "Failed to get project root path: "+err.Error())
		} else {
			hlsDir := filepath.Join(rootPath, "hls", uuid)
			err = os.RemoveAll(hlsDir)
			if err != nil {
				l.logger.Error(context.TODO(), "Failed to delete HLS directory: "+err.Error())
			} else {
				l.logger.Info(context.TODO(), "Deleted HLS directory: "+hlsDir)
			}
		}
		l.logger.Info(context.TODO(), "Closed livestream: " + stream.name)
	} else {
		l.logger.Warn(context.TODO(), "No livestream found with uuid: " + uuid)
	}
	return nil
}

func (l *LivestreamService) UpdateStreamOutPutPathUUID(uuid, outputPathUUID string) error {
	if stream, exists := l.streams[uuid]; exists {
		stream.conn.Close()
		stream.outputPathUUID = outputPathUUID
		l.streams[uuid] = stream
		return nil
	}
	return errors.ErrNotFound
}
