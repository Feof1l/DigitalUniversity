# Проект
Этот проект представляет собой приложение цифрового вуза в рамках хакатона VK & Max.

## Общая архитектура системы


## Запуск проекта
Для запуска проекта необходимо выполнить следующие шаги:

### 1. Docker Compose
Запустите Docker Compose командой:
```bash
docker-compose up --build
```


# Структура базы данных
## Таблица пользователей (users)
```sql
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(100) UNIQUE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role_id INT REFERENCES roles(role_id),
    group_id INT REFERENCES groups(group_id)
);
```

## Таблица ролей (roles)
```sql
CREATE TABLE roles (
    role_id SERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL 
);
```

## Таблица дисциплин (subjects)
```sql
CREATE TABLE subjects (
    subject_id SERIAL PRIMARY KEY,
    subject_name VARCHAR(255) NOT NULL,
    teacher_id INT NOT NULL REFERENCES users(user_id)
);
```

## Таблица предпочтений групп (groups)
```sql
CREATE TABLE groups (
    group_id SERIAL PRIMARY KEY,
    group_name VARCHAR(100) UNIQUE NOT NULL
);
```

## Таблица связи групп и дисциплин (groups_subjects)
```sql
CREATE TABLE groups_subjects (
    group_id INT NOT NULL REFERENCES groups(group_id),
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    PRIMARY KEY (group_id, subject_id)
);
```


## Таблица типов занятий (lesson_types)
```sql
CREATE TABLE lesson_types (
    lesson_type_id SERIAL PRIMARY KEY,
    type_name VARCHAR(50) UNIQUE NOT NULL 
);

```

## Таблица расписания (schedule)
```sql
CREATE TABLE schedule (
    schedule_id SERIAL PRIMARY KEY,
    weekday SMALLINT NOT NULL CHECK (weekday BETWEEN 1 AND 7), 
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    teacher_id INT NOT NULL REFERENCES users(user_id),
    group_id INT NOT NULL REFERENCES groups(group_id),
    lesson_type_id INT NOT NULL REFERENCES lesson_types(lesson_type_id)
);
```

## Таблица оценок (grades)
```sql
CREATE TABLE grades (
    grade_id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES users(user_id),
    teacher_id INT NOT NULL REFERENCES users(user_id),
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    schedule_id INT NOT NULL REFERENCES schedule(schedule_id),
    grade_value INT NOT NULL CHECK (grade_value BETWEEN 0 AND 5),
    grade_date TIMESTAMP DEFAULT NOW()
);
```

## Таблица посещаемости (attendance)
```sql
CREATE TABLE attendance (
    attendance_id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES users(user_id),
    schedule_id INT NOT NULL REFERENCES schedule(schedule_id),
    attended BOOLEAN NOT NULL,
    mark_time TIMESTAMP DEFAULT NOW()
);
```

## Таблица материалов (materials)
```sql
CREATE TABLE materials (
    material_id SERIAL PRIMARY KEY,
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    file_url TEXT NOT NULL,
    uploaded_at TIMESTAMP DEFAULT NOW()
);
```



## Взаимосвязи между таблицами
![alt text](image.png)

## Разработка


### Backend

- Расположен в `src/`
- Модульная структура:
  * `src/application/` - код приложения
  * `src/config/` - конфигурация приложения
  * `src/database/` - работа с базой данных
  * `src/maxAPI/` - бот,интегрируемый с приложением Max
  * `src/models/` - модели данных


## Очистка Docker окружения
```bash
docker stop $(docker ps -a -q)
docker rm $(docker ps -a -q)
docker rmi $(docker images -q)
docker system prune -a --volumes -f
```


