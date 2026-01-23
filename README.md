# 🚀 Productivity App

Полноценное мобильное приложение для планирования задач, ведения дневника и отслеживания годовых целей с синхронизацией через собственный Backend и интеграцией с Telegram.

<p align="center">
  <img src="assets/Снимок экрана 2026-01-22 в 22.05.18.png" width="250" />
  <img src="assets/Снимок экрана 2026-01-22 в 22.05.54.png" width="250" /> 
  <img src="assets/Снимок экрана 2026-01-22 в 22.06.22.png" width="250" />
</p>

<p align="center">
  <img src="assets/Снимок экрана 2026-01-22 в 22.05.18.png" width="250" />
  <img src="assets/Снимок экрана 2026-01-23 в 16.19.25.png" width="250" /> 
  <img src="assets/Снимок экрана 2026-01-23 в 16.20.51.png" width="250" />
</p>

## 🌟 Особенности

*   **План на день:** Управление задачами, календарь, отметки выполнения.
*   **Дневник:** Записи привязаны к датам, сохранение мыслей.
*   **Цели на год:** Древовидная структура (Группы -> Цели), трекинг прогресса.
*   **Telegram Интеграция:** Бот для шеринга планов и отчетов в личку или каналы.
*   **Безопасность:** Полная система авторизации (JWT Access + Refresh tokens).
*   **Offline-first UI:** Оптимистичные обновления интерфейса на клиенте.

---

## 🛠 Технический стек

### 📱 iOS Client (Swift)
*   **UI:** SwiftUI, MVVM Architecture.
*   **Concurrency:** Swift Concurrency (async/await), Task Groups.
*   **Network:** URLSession, Codable (Custom decoding strategies), Generic API Service.
*   **Security:** Keychain (хранение токенов).
*   **Integration:** Deep Links (для связи с Telegram ботом).

### ⚙️ Backend (Go)
*   **Architecture:** Clean Architecture (Handlers -> Services -> Repositories).
*   **Web Framework:** Gin.
*   **Database:** PostgreSQL + pgx (driver) + Squirrel (Query Builder).
*   **Migrations:** Goose (embedded migrations).
*   **Auth:** JWT (Access/Refresh rotation), Bcrypt.
*   **Concurrency:** Graceful Shutdown, Goroutines для бота.
*   **Infrastructure:** Docker, Docker Compose.

---

## 🏗 Архитектура

### Database Schema
Приложение использует реляционную базу данных PostgreSQL.
*   `Users` — Хранение учетных записей.
*   `DailyTasks` / `DiaryEntries` — Данные, привязанные к дате.
*   `YearlyGoals` / `GoalGroups` — Иерархические цели.
*   `TelegramIntegrations` — Связь аккаунта приложения с Telegram ID.

### API & Services
Бэкенд реализует REST API. Взаимодействие с базой данных построено через SQL (`pgx` и `squirrel`) для максимальной производительности и контроля, без использования тяжелых ORM.
