GOOS=linux go build main.go
zip function.zip main
aws lambda update-function-code --function-name vaccineNotify \
--zip-file fileb://function.zip \
--region us-east-1 \
--profile bgock
