package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/configor"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/boj/redistore.v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
	"log"
	"net/http"
	"os"
	common "test-vehcile-monitoring/common/config"
	"test-vehcile-monitoring/common/logger"
	"test-vehcile-monitoring/common/utils"
	"test-vehcile-monitoring/filelogger"
	"test-vehcile-monitoring/handler"
	"test-vehcile-monitoring/message"
	gorm_middleware "test-vehcile-monitoring/monitoring/gorm"
	"test-vehcile-monitoring/session/sessionstore"
)

var (
	config       *common.Config
	serviceLog   *logrus.Entry
	serviceDB    *gorm.DB
	mongoDBConn  *mongo.Client
	sessionStore *redistore.RediStore
	tokenStore   *sessionstore.ServiceSessionStore
)

var (
	newHashKey = []byte("super-secret-key-YYYYYYYYYYYYYYY")
	newBlckKey = []byte("super-secret-key-ZZZZZZZZZZZZZZZ")
	oldHashKey = []byte("super-secret-key-123456789012345")
	oldBlckKey = []byte("super-secret-key-XXXXXXXXXXXXXXX")
)

func main() {
	log.Println("main start")
	defer log.Println("main closed")

	opaInit()
	loadConfig()

	serviceLog = getLogrusEntry(config.Port)

	utils.SetUpCloseHandler()
	defer utils.ReleaseAllResource()

	makeConnection(serviceLog)
	defer TerminateBeforeDisconnection()

	authRouter := mux.NewRouter().PathPrefix("/api/auth").Subrouter()
	authRouter.NotFoundHandler = http.HandlerFunc(notFound)
	authHandler(authRouter, serviceLog)

	/*middleware := session.GatewayMiddleware{
		Logger:       serviceLog,
		SessionStore: sessionStore,
		TokenStore:   tokenStore,
		ServiceDB:    serviceDB,
		Config:       config,
	}*/

	authRouter.Use(
	//middleware.SessionProxy,  // 실제 세션등을 관리하는 미들웨어
	//middleware.LogProxy,      // 어떤 메서드가 호출되는지 출력하는 미들웨어
	//middleware.TraceLogProxy, // Trace 미들웨어
	//middleware.CacheProxy,
	//http_opentracing.Middleware(opentracing.GlobalTracer(), http_opentracing.GorillaMuxOperationFinder), // Jeager
	)

	// 로그인이 필요한 경로
	//http.Handle("/", handlers.CompressHandler(allowCORS(authRouter, config.Cors)))

	nonAuthRouter := mux.NewRouter().PathPrefix("/api/nonauth").Subrouter()
	nonAuthRouter.NotFoundHandler = http.HandlerFunc(notFound)
	nonAuthHandler(nonAuthRouter, serviceLog)

	nonAuthRouter.Use(
	//middleware.LogProxy,      // 어떤 메서드가 호출되는지 출력하는 미들웨어
	//middleware.TraceLogProxy, // Trace 미들웨어
	//middleware.CacheProxy,
	//http_opentracing.Middleware(opentracing.GlobalTracer(), http_opentracing.GorillaMuxOperationFinder), // Jeager
	)

	// 로그인이 필요 없는 경로
	http.Handle("/api/auth/", allowCORS(authRouter, config.Cors))
	http.Handle("/api/nonauth/", allowCORS(nonAuthRouter, config.Cors))
	//http.Handle("/api/noAuth/public", allowCORS(nonAuthRouter, config.Cors))
	//http.Handle("/api/noAuth/internal", allowCORS(nonAuthRouter, config.Cors))
	//http.Handle("/health-check", allowCORS(healthRouter, config.Cors))
	//http.Handle("/api/noAuth/openapi", allowCORS(openApiRouter, config.Cors))

	port := config.Port
	log.Println(fmt.Sprintf("Server starting on port %v", port))
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./public"))))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%v", port), nil))
}

func TerminateBeforeDisconnection() {
	log.Println("main.go termination before disconnection from component...")
}

func opaInit() {

	// init opa
	/*
		Todo open policy agent
	*/

}

func loadConfig() {
	// Config settings
	config = new(common.Config)
	err := configor.New(&configor.Config{}).Load(&config, "env.toml")
	if err != nil {
		//log.Fatalf("fail to configure [%v]\n", err)
		fmt.Println("error : ", err)
	}

	fmt.Println("Config :", config.Cors)
	fmt.Println("Config :", config.RejectExtension)
	fmt.Println("Config :", config.Referer)
	fmt.Println("Config :", config.ServiceDB)

	if config.ErrorLogFile != "" {
		filelogger.Init(config.ErrorLogFile)
	}

}

func allowCORS(router *mux.Router, cors common.Cors) http.Handler {
	return handlers.CORS(
		handlers.AllowedMethods(cors.Methods),
		handlers.AllowedOrigins(cors.Origins),
		handlers.AllowedHeaders(cors.Headers),
		handlers.AllowCredentials(),
	)(router)
}

func authHandler(router *mux.Router, serviceLog *logrus.Entry) {
	baseHandler := &handler.BaseHandler{
		Logger:              serviceLog,
		Config:              config,
		ServiceDB:           serviceDB,
		ServiceSessionStore: tokenStore,
	}

	vehicleHandler := (*handler.VehicleServiceHandler)(baseHandler)

	router.HandleFunc("/vehicle", vehicleHandler.ListVehicles).Methods(http.MethodGet)
}

func nonAuthHandler(router *mux.Router, serviceLog *logrus.Entry) {

	authhandler := &handler.AuthServiceHandler{
		Logger:              serviceLog,
		Config:              config,
		OauthConfig:         nil,
		ServiceDB:           serviceDB,
		ServiceSessionStore: tokenStore,
	}

	router.HandleFunc("/google/login", authhandler.GoogleLogin).Methods(http.MethodGet)
	router.HandleFunc("/google/callback", authhandler.GoogleAuthCallback).Methods(http.MethodGet)
}

func getLogrusEntry(port string) *logrus.Entry {
	hostName, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	return logger.InitServiceLogger("TEST", "TEST", "HTTP").
		WithField(logger.LogURL, fmt.Sprintf("%s:%s", hostName, port))
}

func makeConnection(serviceLog *logrus.Entry) {
	serviceDB = makeConnectionDB(serviceLog, config.ServiceDB.DriverName, config.ServiceDB.DataSourceName, config.ServiceDB.MaxOpenConns, config.ServiceDB.MaxIdleConns)
	mongoDBConn = mongoDBConnection(serviceLog, config.MongoDB.URL)
	gorm_middleware.SetGlobalGorm(serviceDB)

	sessionStore = makeRedisStoreConnection(serviceLog, 10, "tcp", config.Redis.AuthStore, config.Redis.AuthStorePwd, "0", newHashKey, newBlckKey, oldHashKey, oldBlckKey)
	tokenStore = makeSessionStoreConnection(serviceLog, config.Redis.AuthStore, config.Redis.AuthStorePwd, 1, 10)
}

func makeRedisStoreConnection(service *logrus.Entry, size int, network, address, password, db string, keyPairs ...[]byte) *redistore.RediStore {
	store, err := redistore.NewRediStoreWithDB(size, network, address, password, db, keyPairs...)

	if err != nil {
		serviceLog.Fatal(fmt.Sprintf("unable to init sessionstore %v", err.Error()))
	}

	utils.RegisterResourceCloser(func() {
		if store != nil {
			serviceLog.Infof("Closing redis store connectionm, network=%s, address=%s, db=%s", network, address, db)
			_ = store.Close()
		}
	})

	return store
}

// 세션 스토어 연결하기
func makeSessionStoreConnection(serviceLog *logrus.Entry, host string, pw string, db int, poolSize int) *sessionstore.ServiceSessionStore {
	sessionStore := &sessionstore.ServiceSessionStore{}
	if err := sessionStore.Open(host, pw, db, poolSize); err != nil {
		serviceLog.Fatal(err)
	}

	utils.RegisterResourceCloser(func() {
		if sessionStore != nil {
			serviceLog.Infof("Closing service session store connection, host=%s, db=%d", host, db)
			sessionStore.Close()
		}
	})

	return sessionStore
}

func makeConnectionDB(serviceLog *logrus.Entry, driverName string, databaseSource string, maxConn, maxIdleConn int) *gorm.DB {
	db, err := gorm.Open(mysql.Open(databaseSource), &gorm.Config{
		Logger: gorm_logger.Default.LogMode(gorm_logger.Info),
	})

	if err != nil {
		serviceLog.Fatalf("failed to init DB: %v %s", err, databaseSource)
	}

	return db
}

func mongoDBConnection(serviceLog *logrus.Entry, databaseSource string) *mongo.Client {
	clientOprions := options.Client().ApplyURI(databaseSource)
	client, err := mongo.Connect(context.TODO(), clientOprions)
	if err != nil {
		serviceLog.Fatalf("failed to init DB : %v %s", err, databaseSource)
	}

	if err = client.Ping(context.TODO(), nil); err != nil {
		serviceLog.Fatalf("failed to init DB : %v %s", err, databaseSource)
	}

	serviceLog.Info("MongoDB Connection Made")
	return client
}

func notFound(w http.ResponseWriter, r *http.Request) {
	msg := &message.ErrorMessage{
		ErrorCode: 404,
		Message:   fmt.Sprintf("Page not found\n%s%s", r.Host, r.URL),
	}
	w.WriteHeader(msg.ErrorCode)
	json.NewEncoder(w).Encode(msg)
}
