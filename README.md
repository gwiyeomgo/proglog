
compile1:
protoc api/v1/log.proto --go_out=. --go_opt=paths=source_relative --proto_path=.

compile2:
protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
api/v1/log.proto


## log 라이브러리 
한 사람이 한순간에 한 대의 컴퓨터에서만 사용 가능
그 사용자는 라이브러리의 api를 배우고 코드를 실행하며 자신의 디스크에 로그를 저장

##  grpc 로그 클라이언트 (grpc 서버 정의,grpc protobuf 컴파일하여 코드 생성)  
로그라이브러리를 기반으로 여러 사람이 같은 데이터로 소통하는 서비스
grpc 는 보안 소켓 계층과 전송 계층 보안을 지원하여 클라이언트와 서버 사이를 오가는 모든 데이터를 암호화한다


grpc 란 관련이 있는 rpc 엔드포인트들을 묶은 그룹

컴파일러가 grpc 로그 클라이언트를 만들고
로그 서비스 api 를 구현할 준비 완료
```
protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        api/v1/log.proto
```

### grpc 서버 생성

### 서비스 보안의 세단계
    * 주고받는 데이터 암호화(중간자 공격 ex)도청)
        * 중간자 공격으로부터 막아주는,가장 널리 쓰이는 암호학 방법 TLS
        * TLS 는 단방향 인증으로 서버만 인증
        
> go install github.com/cloudflare/cfssl/cmd/cfssl@latest

> go install github.com/cloudflare/cfssl/cmd/cfssljson@latest
    ca-csr.json : ca 인증서 설정
    ca-config.json : ca가 어떤 인증서를 발행할지 설정(정책 설정)
    server-csr.json : 서버 인증서 설정


    * 클라이언트 인증 = 클라이언트가 누군지 확인
        * 애플리케이션에서 사용자명,비밀번호와 토큰의 조합으로 구현
    * 인증한 클라이언트의 권한을 결정
        인증와 권한결정(인가)


https://github.com/casbin/casbin

### 환경변수
HOME

CONFIG_PATH : 인증서를 저장할 위치
CONFIG_DIR : 설정 파일들의 경로
ex) ./test