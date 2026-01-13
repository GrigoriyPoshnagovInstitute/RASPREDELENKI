#### 1. Полный запуск с нуля (Install)
Если ты перезагрузил комп или хочешь все почистить:

```bash
./install.sh
```
*Этот скрипт сам всё удалит, скачает, исправит баги, загрузит и дождется статуса Running.*

#### 2. Включение доступа (Ports)
Когда `install.sh` написал "ВСЕ ГОТОВО", или если ты просто хочешь подключиться к уже работающему кластеру:

```bash
./ports.sh
```
*Этот скрипт будет висеть в терминале. Не закрывай его.*

#### 3. Сдача (Scenario)
В **новом** терминале (пока `ports.sh` работает в соседнем):

1.  **Создать богача:**
    ```bash
    grpcurl -plaintext -d '{"user": {"login": "rich_guy", "status": 1, "email": "rich@test.com"}}' localhost:8081 User.UserInternalService/StoreUser
    ```
2.  **Закинуть денег:**
    ```bash
    grpcurl -plaintext -d '{"balance": {"userID": "ID_ОТ_БОГАЧА", "balance": 100000}}' localhost:8085 Payment.PaymentInternalService/StoreUserBalance
    ```
3.  **Создать товар:**
    ```bash
    grpcurl -plaintext -d '{"product": {"name": "MacBook", "price": 50000}}' localhost:8083 Product.ProductInternalService/StoreProduct
    ```
4.  **Купить:**
    ```bash
    grpcurl -plaintext -d '{"userID": "ID_ОТ_БОГАЧА", "items": [{"productID": "ID_ТОВАРА", "quantity": 1}]}' localhost:8087 Order.OrderInternalService/CreateOrder
    ```
5.  **Смотреть:** [http://localhost:8080](http://localhost:8080)



Поды приложения

```bash
kubectl get pods -n application -
```

Поды инфраструктуры

```bash
kubectl get pods -n infrastructure -
```