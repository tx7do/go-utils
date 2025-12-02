echo on

::指定起始文件夹
set DIR="%cd%"

go get all
go mod tidy

cd %DIR%/bank_card
go get all
go mod tidy

cd %DIR%/geoip
go get all
go mod tidy

cd %DIR%/copierutil
go get all
go mod tidy

cd %DIR%/translator
go get all
go mod tidy

cd %DIR%/entgo
go get all
go mod tidy

cd %DIR%/gorm
go get all
go mod tidy