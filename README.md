# 모바일 청첩장 서버

모바일 청첩장 웹 애플리케이션([프런트 저장소](https://github.com/devmtn30/wedding-invitation))과 페어로 동작하는 백엔드입니다.  
GCP Cloud run으로 배포하기 때문에 재배포 시 데이터가 소실되는 SQLite 대신 **Firestore**를 기본 저장소로 사용하며, Cloud Run 배포를 염두에 둔 구성으로 되어 있습니다.  
관리자가 참석/방명록 데이터를 직접 확인하고 CRUD 할 수 있도록 `/admin` 머테리얼 UI 콘솔과 백업/복원 API도 제공됩니다.

origin source: https://github.com/juhonamnam/wedding-invitation-server.git

## 요구 사항

- Go ≥ 1.21
- (선택) Cloud SDK & Docker: Cloud Run 배포/이미지 빌드를 위해 필요
- Firestore 프로젝트 및 접근 권한 (로컬 실행 시 Application Default Credentials 또는 서비스 계정 키 필요)

## 빠른 시작

1. 저장소 복제

   ```bash
   git clone https://github.com/juhonamnam/wedding-invitation-server.git
   cd wedding-invitation-server
   ```

2. 의존성 다운로드

   ```bash
   go mod download
   ```

3. 환경 변수 설정  
   샘플은 `.env.example` 에 있습니다. 필요한 값을 복사해 사용하거나, 직접 환경 변수로 지정하세요.

   | 키 | 설명 |
   | --- | --- |
   | `GCP_PROJECT_ID` | Firestore가 속한 GCP 프로젝트 ID (**필수**) |
   | `ADMIN_PASSWORD` | `/admin` 콘솔과 관리자 API 호출 시 사용할 비밀번호 |
   | `ALLOW_ORIGIN` | CORS 허용 도메인 배열 (주로 프런트엔드 URL) |

   로컬에서 서비스 계정 키를 사용할 때는 `GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json` 도 함께 설정해야 합니다.

4. 서버 실행

   ```bash
   go run app.go
   ```

   기본 포트는 `8080` 입니다.

## 관리자 콘솔 & 데이터 백업

- `http://localhost:8080/admin` (배포 환경에서는 Cloud Run 서비스 URL 뒤에 `/admin`)  
  - 상단에 `Admin Password` 입력 후 **Apply** 를 눌러 인증하면 방명록/참석 정보를 실시간으로 조회하고 CRUD 할 수 있습니다.  
  - 콘솔은 Material Design Web Components 로 렌더링되며, 브라우저 로컬에서 API 호출 시 `X-Admin-Password` 헤더를 자동으로 첨부합니다.

- **백업/복원 API**
  - `POST /admin/import`  
    ```json
    {
      "adminPassword": "비밀번호",
      "guestbook": { "posts": [ ... ], "total": 0 },
      "attendance": [ ... ]
    }
    ```
    기존 SQLite 백업을 Firestore로 옮길 때 사용합니다.
  - `GET /admin/api/guestbook?offset=0&limit=100`
  - `GET /admin/api/attendance`  
    관리자 비밀번호를 `X-Admin-Password` 헤더로 전달하면 JSON 덤프를 받을 수 있으므로 주기적인 백업에 활용하세요.

## Public API

| 경로 | 메서드 | 설명 |
| --- | --- | --- |
| `/guestbook` | `GET` | `offset`, `limit` 쿼리로 공개 방명록 목록 조회 |
|  | `POST` | `{ "name", "content", "password" }` 로 방명록 작성 |
|  | `PUT` | `{ "id", "password" }` 로 글 비활성화(삭제) |
| `/attendance` | `POST` | `{ "side", "name", "meal", "count" }` 로 참석 의사 기록 |
|  | `GET` | 모든 참석 데이터를 최신순으로 반환 |

Admin 전용 API (`/admin/api/...`) 는 콘솔이 사용하는 내부 엔드포인트입니다. 직접 호출하려면 `X-Admin-Password` 헤더가 필요합니다.

## Firestore 인덱스

아래 복합 인덱스는 반드시 생성되어야 합니다. (콘솔에서 쿼리 에러 링크로 바로 만들 수 있습니다.)

| 컬렉션 | 필드 | 정렬 |
| --- | --- | --- |
| `guestbook` | `valid` | Ascending |
|  | `timestamp` | Descending |

참석 컬렉션(`attendance`)은 기본 단일 필드 정렬만 사용합니다.

## Cloud Run 배포

1. 컨테이너 이미지 빌드 & Push

   ```bash
   docker buildx build \
     --platform linux/amd64 \
     -t ${IMAGE}:${TAG} \
     --push .
   ```

2. Cloud Run 배포 (필수 환경 변수 지정)

   ```bash
   gcloud run deploy wedding-invitation-server \
     --image ${IMAGE}:${TAG} \
     --region asia-northeast3 \
     --platform managed \
     --allow-unauthenticated \
     --timeout=600s \
     --cpu=1 \
     --memory=512Mi \
     --set-env-vars GCP_PROJECT_ID=${PROJECT_ID},ADMIN_PASSWORD=${ADMIN_PASSWORD},ALLOW_ORIGIN=${ALLOW_ORIGIN} \
     --project ${PROJECT_ID}
   ```

   Cloud Run 서비스 계정에는 Firestore 접근 권한(예: `roles/datastore.user`)을 부여해야 합니다.  
   배포 후 `https://<SERVICE_URL>/admin` 에 접속해 관리자 콘솔이 정상 동작하는지 확인하세요.
