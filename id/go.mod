module github.com/tx7do/go-utils/id

go 1.24.0

require (
	github.com/bwmarrin/snowflake v0.3.0
	github.com/google/uuid v1.6.0
	github.com/lithammer/shortuuid/v4 v4.2.0
	github.com/rs/xid v1.6.0
	github.com/segmentio/ksuid v1.0.4
	github.com/sony/sonyflake v1.3.0
	github.com/stretchr/testify v1.11.1
	github.com/tx7do/go-utils v1.1.34
	go.mongodb.org/mongo-driver v1.17.7
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/tx7do/go-utils => ../
