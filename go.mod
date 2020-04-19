module github.com/generalzgd/grpc-gw

go 1.14

require (
	github.com/armon/go-metrics v0.3.3 // indirect
	github.com/astaxie/beego v1.12.1
	github.com/funny/slab v0.0.0-20180511031532-b1fad5e5d478
	github.com/generalzgd/comm-libs v0.0.0-20200419072240-8400406771df
	github.com/generalzgd/deepcopy v0.0.0-20200412030653-dddc3fe863a1
	github.com/generalzgd/grpc-svr-frame v0.0.0-20191018022130-7aae0a1a0a61
	github.com/generalzgd/link v0.0.0-20190924020608-c636ec1cdbe5
	github.com/generalzgd/micro-cmdid v0.0.0-20200415163720-e4affed55ddc
	github.com/generalzgd/micro-proto v0.0.0-20200415163540-f55ea058bc63
	github.com/golang/protobuf v1.3.3
	github.com/gorilla/websocket v1.4.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.2
	github.com/toolkits/slice v0.0.0-20141116085117-e44a80af2484
	google.golang.org/grpc v1.24.0
)

replace (
	cloud.google.com/go => github.com/googleapis/google-cloud-go v0.37.4
	github.com/generalzgd/comm-libs => ../comm-libs
	github.com/generalzgd/grpc-svr-frame => ../grpc-svr-frame
	github.com/generalzgd/micro-cmdid => ../micro-cmdid
	github.com/generalzgd/micro-proto => ../micro-proto
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190513172903-22d7a77e9e5f
	golang.org/x/exp => github.com/golang/exp v0.0.0-20190718202018-cfdd5522f6f6
	golang.org/x/image => github.com/golang/image v0.0.0-20190703141733-d6a02ce849c9
	golang.org/x/lint => github.com/golang/lint v0.0.0-20190409202823-959b441ac422
	golang.org/x/mobile => github.com/golang/mobile v0.0.0-20190719004257-d2bd2a29d028
	golang.org/x/mod => github.com/golang/mod v0.1.0
	golang.org/x/net => github.com/golang/net v0.0.0-20190827160401-ba9fcec4b297
	golang.org/x/oauth2 => github.com/golang/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync => github.com/golang/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190712062909-fae7ac547cb7
	golang.org/x/text => github.com/golang/text v0.3.2
	golang.org/x/time => github.com/golang/time v0.0.0-20190308202827-9d24e82272b4
	golang.org/x/tools => github.com/golang/tools v0.0.0-20191217033636-bbbf87ae2631
	golang.org/x/xerrors => github.com/golang/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/api v0.3.1 => github.com/googleapis/google-api-go-client v0.3.1
	google.golang.org/appengine => github.com/golang/appengine v1.6.1
	google.golang.org/genproto => github.com/googleapis/go-genproto v0.0.0-20190516172635-bb713bdc0e52
	google.golang.org/grpc => github.com/grpc/grpc-go v1.24.0
)
