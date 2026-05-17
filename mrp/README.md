# Zeus MRP Service

Hãy dựa trên layout được community chấp nhận: [golang-standards/project-layout](https://github.com/golang-standards/project-layout)

Dưới đây là cấu trúc đơn giản nhưng hiệu quả (Simple & Effective) cho một service nhỏ:

```text
zeus-mrp-service/
├── internal/
│   ├── controllers/
│   │   ├── controllers.go              # Shared controller struct & constructor
│   │   ├── create_order.go             # Handler: POST /production/orders
│   │   └── get_shortages.go            # Handler: GET /mrp/shortages
│   ├── service/
│   │   ├── service.go                  # Shared service struct & constructor
│   │   ├── plan_production.go          # Logic: Planning
│   │   ├── plan_production_test.go     # [Unit Test] for Planning
│   │   ├── run_bom_explosion.go        # Logic: BOM Explosion
│   │   ├── run_bom_explosion_test.go   # [Unit Test] for BOM Explosion
│   │   ├── check_ctb.go                # Logic: Clear-to-Build check
│   │   └── check_ctb_test.go           # [Unit Test] for CTB check
│   ├── repository/
│   │   ├── mrp_repository.go
│   │   ├── cache_repository.go
│   │   └── sqlite/
│   │       └── mrp_sqlite.go           # SQLite Implementation
│   └── models/
├── tests/
└── README.md
```

### Giải thích các thành phần (Components Explanation):

1.  **`cmd/`**: Chỉ chứa code khởi tạo (boilerplate). Không viết logic ở đây.
2.  **`internal/models/`**: Để giữ mọi thứ đơn giản, bạn có thể gộp **Entities** (Thực thể) và **Models** vào đây.
3.  **`internal/repository/`**: Tách biệt (Decoupling) logic truy vấn database khỏi logic nghiệp vụ.
4.  **`internal/service/`**: Đây là "trái tim" của ứng dụng, nơi điều hướng dữ liệu từ Repository sang Controller.
5.  **`tests/`**: Chứa các bài kiểm tra tích hợp (Integration tests) và dữ liệu test.

---

### Cấu trúc Testing (Testing Strategy):

Để đảm bảo chất lượng code, chúng ta tuân thủ quy tắc sau:

#### 1. Unit Tests (Kiểm tra đơn vị)
*   **Vị trí**: Nằm cùng folder với file code cần test.
*   **Tên file**: Kết thúc bằng `_test.go`.
*   **Mục đích**: Kiểm tra các logic nhỏ, độc lập trong một function hoặc struct.
*   **Ví dụ**: `internal/service/calculator.go` sẽ có file test là `internal/service/calculator_test.go`.

#### 2. Integration Tests (Kiểm tra tích hợp)
*   **Vị trí**: Nằm trong thư mục `tests/integration/`.
*   **Mục đích**: Kiểm tra sự phối hợp giữa nhiều package (ví dụ: Controller -> Service -> Repository) hoặc kiểm tra với Database thật.
*   **Ví dụ**: `tests/integration/user_flow_test.go`.

**Từ vựng kỹ thuật (Keywords):**
*   **Decoupling**: Sự tách biệt giữa các thành phần để dễ bảo trì và mở rộng.
*   **Entry point**: Điểm xuất phát của chương trình.
*   **Boilerplate**: Các đoạn mã lặp đi lặp lại cần thiết để khởi chạy dự án.
*   **Co-located**: Đặt các file liên quan ở cùng một chỗ (áp dụng cho Unit Test).
 