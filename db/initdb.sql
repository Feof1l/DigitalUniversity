CREATE TABLE roles (
    role_id SERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL
);
CREATE TABLE groups (
    group_id SERIAL PRIMARY KEY,
    group_name VARCHAR(100) UNIQUE NOT NULL
);
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    username VARCHAR(100) UNIQUE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role_id INT REFERENCES roles(role_id),
    group_id INT REFERENCES groups(group_id)
);
CREATE TABLE subjects (
    subject_id SERIAL PRIMARY KEY,
    subject_name VARCHAR(255) NOT NULL,
    teacher_id INT NOT NULL REFERENCES users(user_id)
);
CREATE TABLE groups_subjects (
    group_id INT NOT NULL REFERENCES groups(group_id),
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    PRIMARY KEY (group_id, subject_id)
);
CREATE TABLE lesson_types (
    lesson_type_id SERIAL PRIMARY KEY,
    type_name VARCHAR(50) UNIQUE NOT NULL
);
CREATE TABLE schedule (
    schedule_id SERIAL PRIMARY KEY,
    weekday SMALLINT NOT NULL CHECK (
        weekday BETWEEN 1 AND 6
    ),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    teacher_id INT NOT NULL REFERENCES users(user_id),
    group_id INT NOT NULL REFERENCES groups(group_id),
    lesson_type_id INT NOT NULL REFERENCES lesson_types(lesson_type_id)
);
CREATE TABLE grades (
    grade_id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES users(user_id),
    teacher_id INT NOT NULL REFERENCES users(user_id),
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    schedule_id INT NOT NULL REFERENCES schedule(schedule_id),
    grade_value INT NOT NULL CHECK (
        grade_value BETWEEN 0 AND 5
    ),
    grade_date TIMESTAMP DEFAULT NOW()
);
CREATE TABLE attendance (
    attendance_id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES users(user_id),
    schedule_id INT NOT NULL REFERENCES schedule(schedule_id),
    attended BOOLEAN NOT NULL,
    mark_time TIMESTAMP DEFAULT NOW()
);
CREATE TABLE materials (
    material_id SERIAL PRIMARY KEY,
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    file_url TEXT NOT NULL,
    uploaded_at TIMESTAMP DEFAULT NOW()
);
