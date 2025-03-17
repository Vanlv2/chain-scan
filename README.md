# Hướng Dẫn Triển Khai Bot WhatsApp API

Hướng dẫn cách đăng ký, cấu hình và triển khai bot sử dụng WhatsApp Business API với Webhook.

## 1. Giới thiệu

WhatsApp Business API cho phép doanh nghiệp giao tiếp với khách hàng qua WhatsApp một cách tự động và hiệu quả. có thể sử dụng API để gửi tin nhắn, nhận phản hồi và quản lý hội thoại từ khách hàng thông qua webhook.

## 2. Yêu cầu hệ thống

Trước khi bắt đầu, cần chuẩn bị:

- Tài khoản Facebook Developer.
- Một ứng dụng Facebook được liên kết với WhatsApp Business API.
- Máy chủ để lắng nghe các webhook từ WhatsApp (có thể là máy chủ bất kỳ có thể kết nối internet).
- Token truy cập API từ Facebook Developer.

## 3. Chuẩn bị

Để bắt đầu, cần tạo tài khoản Facebook Developer và một ứng dụng WhatsApp:

1. Truy cập: [https://developers.facebook.com/apps/creation/](https://developers.facebook.com/apps/creation/).
2. Tạo một ứng dụng mới và chọn loại **Business**.
3. Sau khi tạo, sẽ có **App ID** và **App Secret**. Thông tin này sẽ được sử dụng để cấu hình kết nối với WhatsApp API.

## 4. Đăng ký WhatsApp Business API

1. Vào trang quản lý ứng dụng của Facebook Developer.
2. Truy cập mục **WhatsApp > Getting Started**.
3. Kết nối hoặc tạo tài khoản WhatsApp Business của bạn.
4. Cấu hình số điện thoại mà bot sẽ sử dụng để nhận và gửi tin nhắn.
5. sẽ nhận được một **Token**. Token này sẽ được dùng để thực hiện các yêu cầu API.

## 5. Cấu hình webhook

Webhook sẽ cho phép ứng dụng của nhận tin nhắn từ người dùng WhatsApp.

### Bước 1: Thiết lập webhook

1. Truy cập **Webhooks** trên trang quản lý ứng dụng Facebook Developer.
2. Cung cấp URL của server Webhook (ví dụ: `https://your-server.com/webhook`).
3. Xác thực webhook bằng cách cung cấp `verify_token` mà đã chọn.
4. Chọn các loại sự kiện muốn nhận, bao gồm tin nhắn WhatsApp và thông báo trạng thái.

### Bước 2: Xác thực webhook

Khi cấu hình webhook, Facebook sẽ gửi một yêu cầu GET để xác thực. Đảm bảo rằng máy chủ của phản hồi yêu cầu theo định dạng sau:

```bash
GET /webhook?hub.verify_token=your_token&hub.challenge=CHALLENGE_ACCEPTED&hub.mode=subscribe

## 6. Gửi và nhận tin nhắn

```bash
POST https://graph.facebook.com/v14.0/<Phone-Number-ID>/messages
Authorization: Bearer <Token>
Content-Type: application/json

{
  "messaging_product": "whatsapp",
  "to": "<User-Phone-Number>",
  "type": "text",
  "text": {
    "body": "Xin chào! Đây là tin nhắn từ bot WhatsApp."
  }
}

## Lưu ý quan trọng
Quyền truy cập: Kiểm tra kỹ quyền truy cập của ứng dụng trên Facebook Developer để đảm bảo nó có đủ quyền để truy cập API WhatsApp.
Xử lý lỗi: Hãy đảm bảo rằng hệ thống của bạn xử lý được các trường hợp như token hết hạn, lỗi xác thực webhook, hoặc lỗi kết nối API.
Chạy thử: Sau khi hoàn thành cấu hình, chạy thử bot WhatsApp bằng cách gửi tin nhắn từ tài khoản WhatsApp cá nhân đến số điện thoại đã đăng ký.
