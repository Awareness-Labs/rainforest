# Architecture

![](/docs/img/arch.jpg)

Rainforest 在架構上的設計屬於聯邦式的 Hub-Leaf 結構 (or Hub-Spoke)，Runtime 種類分為以下：

- Hub：負責整合和暴露連線在此 Hub 的所有 Leaf
- Leaf：使用者實際服務的 Runtime

 Leaf 元件中有以下幾個組成：

- Data Product API：封裝 JetStream 的一組 API，用於建立實體 Data Product
- Adapter API：Adapter 原理上只是一個 Stateful Consumer，所以可以直接建立在 Data Product 上面，主要設計參考 **[DBLog: A Watermark Based Change-Data-Capture Framework](https://arxiv.org/abs/2010.12597)**。
- Stream Engine：JetStream 本身，Data Product API 會直接轉換成 Stream
    - State：Subject Limit 設定為 1，並且執行 LWW 的協議。
    - Event：正常人認知的 Event Store
- Embedded OLTP：用 rocksdb 的近親，BadgerDB 實作，擁有 SSI 等級的 Isolation
- Embedded OLAP：DuckDB 實例

# Project Philosophy

以下是 Rainforest 主張的價值 (絕對不會妥協的價值)：

- 輕量化的 Data Mesh Runtime
    - Rainforest Code-base 最小化
    - Image 最小化
    - 部屬最小化 (能夠以一個 Data Product 為單位部屬，而且避免掛尿袋的設計)
- 聯邦式的網路架構絕對不能更改 (例如改成 Shard 為基礎，但是是統一部屬的系統)
- 簡單、穩定、可擴充 (不論是 Publish/Consume 都該變成 Framework)
- Queue-based API


# Information Talk to Brobridge

- President Example
- IT Home/InfoQ/MIT 6.824
- Striim / Postgres / DBLog
- We are information RD vs Infra (NATS Blog Example)
- Information Order (Digital Brain Example) / Logic Pre-Experience
- i Phone 通知 Example (R&D + Frontend)
- Belief Value (Skill vs Corp Income)
- Communication is not exist we are playing LANG GAME (蒙古部落)