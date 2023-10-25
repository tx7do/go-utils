echo off

::指定起始文件夹
set DIR="%cd%"

go get all
go mod tidy

cd %DIR%/bank_card
go get all
go mod tidy

cd %DIR%/entgo
go get all
go mod tidy

cd %DIR%/geoip
go get all
go mod tidy
