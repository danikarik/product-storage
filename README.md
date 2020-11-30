# Task Description

Требуется написать gRPC-сервер на языке GoLang (1.13+), с постоянным хранилищем
MongoDB, реализующий 2 метода:
- Fetch(URL) - запросить внешний CSV-файл со списком продуктов по внешнему адресу.
CSV-файл имеет вид PRODUCT NAME;PRICE. Последняя цена каждого продукта должна
быть сохранена в базе с датой запроса. Также нужно сохранять количество изменений
цены продукта.
- List(<paging params>, <sorting params>) - получить постраничный список продуктов с их
ценами, количеством изменений цены и датами их последнего обновления.
Предусмотреть все варианты сортировки для реализации интерфейса в виде
бесконечного скролла.

Сервер должен быть запущен в 2+ экземплярах (каждый в своем Docker-контейнере) и
закрыт балансировщиком, соответствующие конфигурации также должны быть
предоставлены для тестовой среды.
