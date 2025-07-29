@echo off
if not exist bin mkdir bin
go build -o bin\excel-schema-generator.exe .