### Шаг 1: Проброс портов (Port Forwarding)

Так как сервисы живут внутри кластера, мы не можем достучаться до них напрямую. Нужно открыть "туннели".

Открой **5 разных терминалов** (или вкладок) и запусти в каждом по одной команде. **Не закрывай их**, пока проверяешь работу.

1.  **Temporal UI** (чтобы видеть красивые графики):
    ```bash
    kubectl port-forward svc/temporal-ui -n infrastructure 8080:8080
    ```
2.  **User Service** (Порт 8081):
    ```bash
    kubectl port-forward svc/userservice -n application 8081:8081
    ```
3.  **Product Service** (Порт 8083 -> 8081 внутри):
    ```bash
    kubectl port-forward svc/productservice -n application 8083:8081
    ```
4.  **Payment Service** (Порт 8085 -> 8081 внутри):
    ```bash
    kubectl port-forward svc/paymentservice -n application 8085:8081
    ```
5.  **Order Service** (Порт 8087 -> 8081 внутри):
    ```bash
    kubectl port-forward svc/orderservice -n application 8087:8081
    ```

---

### Шаг 2: Прогон Сценария (Happy Path)

Теперь, когда порты открыты, ты можешь выполнять команды `grpcurl` прямо из методички в **шестом терминале**.

**1. Открой Temporal UI в браузере:** [http://localhost:8080](http://localhost:8080)
Там пока должно быть пусто.

**2. Создай Богатого Студента:**
```bash
grpcurl -plaintext -d '{"user": {"login": "rich_student", "status": 1, "email": "rich@test.com"}}' localhost:8081 User.UserInternalService/StoreUser
```
👉 **Скопируй `userID` из ответа!**

**3. Начисли ему денег (PaymentService):**
Вставь полученный `userID` вместо `RICH_ID`:
```bash
grpcurl -plaintext -d '{"balance": {"userID": "RICH_ID", "balance": 100000}}' localhost:8085 Payment.PaymentInternalService/StoreUserBalance
```

**4. Создай Товар (ProductService):**
```bash
grpcurl -plaintext -d '{"product": {"name": "MacBook Pro", "price": 50000}}' localhost:8083 Product.ProductInternalService/StoreProduct
```
👉 **Скопируй `productID` из ответа!**

**5. Подожди 5-10 секунд** (чтобы сообщения через RabbitMQ дошли до OrderService).

**6. Создай Заказ (OrderService -> Temporal):**
Вставь свои ID:
```bash
grpcurl -plaintext -d '{"userID": "RICH_ID", "items": [{"productID": "PROD_ID", "quantity": 1}]}' localhost:8087 Order.OrderInternalService/CreateOrder
```

---

### Шаг 3: Проверка результата

1.  Вернись в браузер: [http://localhost:8080](http://localhost:8080).
2.  Обнови страницу. Ты должен увидеть Workflow со статусом **Completed** (зеленый).
3.  Кликни на него — ты увидишь историю выполнения:
    *   Activity `ReserveProducts` — выполнено.
    *   Activity `ProcessPayment` — выполнено.
    *   Activity `SendOrderCreatedNotification` — выполнено.

Если ты это видишь — **лабораторная сдана**. Система работает в Kubernetes как часы.