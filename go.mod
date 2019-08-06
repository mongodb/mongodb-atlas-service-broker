module github.com/mongodb/mongodb-atlas-service-broker

go 1.11

require (
	code.cloudfoundry.org/lager v2.0.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/kubernetes-incubator/service-catalog v0.2.1
	github.com/kubernetes-sigs/service-catalog v0.2.1
	github.com/pivotal-cf/brokerapi v5.1.0+incompatible
	github.com/pkg/errors v0.8.1 // indirect
	github.com/stretchr/testify v1.3.0
	go.mongodb.org/mongo-driver v1.0.4
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	k8s.io/api v0.0.0-20190806064354-8b51d7113622
	k8s.io/apimachinery v0.0.0-20190802060556-6fa4771c83b3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20190801114015-581e00157fb1 // indirect
)
