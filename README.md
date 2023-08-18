



## log 라이브러리
한 사람이 한순간에 한 대의 컴퓨터에서만 사용 가능
그 사용자는 라이브러리의 api를 배우고 코드를 실행하며 자신의 디스크에 로그를 저장
##  grpc 로그 클라이언트
로그라이브러리를 기반으로 여러 사람이 같은 데이터로 소통하는 서비스
grpc 는 보안 소켓 계층과 전송 계층 보안을 지원하여 클라이언트와 서버 사이를 오가는 모든 데이터를 암호화한다


grpc 란 관련이 있는 rpc 엔드포인트들을 묶은 그룹

컴파일러가 grpc 로그 클라이언트를 만들고
로그 서비스 api 를 구현할 준비 완료
```go
protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        api/v1/log.proto
```

###  grpc 서버
