package photogalery

// type loggingService struct {
// 	logger log.Logger
// 	Service
// }

// // NewLoggingService returns a new instance of a logging Service.
// func NewLoggingService(logger log.Logger, s Service) Service {
// 	return &loggingService{logger, s}
// }

// func (s *loggingService) Upload(file *multipart.FileHeader) error {
// 	defer func(begin time.Time) {
// 		s.logger.Log(
// 			"method", "upload",
// 			"took", time.Since(begin),
// 		)
// 	}(time.Now())
// 	return s.Service.Upload(file)
// }
