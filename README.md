# Medods test task

## Содержание
- [Технологии](#технологии)
- [Использование](#использование)
- [Тестирование](#тестирование)
- [To do](#to-do)

## Технологии
- Go
- PostgreSQL
- Docker

## Использование

### Требования:
- [Task](https://taskfile.dev/)
- [Docker](https://www.docker.com/)

#### Для тестов(только локально):
- [Golang](https://go.dev/)

### Запуск
```sh
task run
```

### Остановка
```sh
task stop
```

## Тестирование
```sh
task run_unit_tests
```
```sh
task run_e2e_tests
```

## To do
- [x] Добавить крутое README
- [ ] Убрать слипы в e2e тестах
- [ ] Всё переписать
