CREATE TABLE IF NOT EXISTS roles (
    role_id SERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL
);
CREATE TABLE IF NOT EXISTS groups (
    group_id SERIAL PRIMARY KEY,
    group_name VARCHAR(100) UNIQUE NOT NULL
);
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    usermax_id BIGINT UNIQUE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role_id INT REFERENCES roles(role_id),
    group_id INT REFERENCES groups(group_id)
);
CREATE TABLE IF NOT EXISTS subjects (
    subject_id SERIAL PRIMARY KEY,
    subject_name VARCHAR(255) NOT NULL,
    teacher_id INT NOT NULL REFERENCES users(user_id)
);
CREATE TABLE IF NOT EXISTS groups_subjects (
    group_id INT NOT NULL REFERENCES groups(group_id),
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    PRIMARY KEY (group_id, subject_id)
);
CREATE TABLE IF NOT EXISTS lesson_types (
    lesson_type_id SERIAL PRIMARY KEY,
    type_name VARCHAR(50) UNIQUE NOT NULL
);
CREATE TABLE IF NOT EXISTS schedule (
    schedule_id SERIAL PRIMARY KEY,
    weekday SMALLINT NOT NULL CHECK (
        weekday BETWEEN 1 AND 7
    ),
    class_room VARCHAR(100),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    subject_id INT NOT NULL REFERENCES subjects(subject_id),
    teacher_id INT NOT NULL REFERENCES users(user_id),
    group_id INT NOT NULL REFERENCES groups(group_id),
    lesson_type_id INT NOT NULL REFERENCES lesson_types(lesson_type_id)
);
CREATE TABLE IF NOT EXISTS grades (
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
CREATE TABLE IF NOT EXISTS attendance (
    attendance_id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES users(user_id),
    schedule_id INT NOT NULL REFERENCES schedule(schedule_id),
    attended BOOLEAN NOT NULL,
    mark_time TIMESTAMP DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_teacher_unique ON users(first_name, last_name, role_id)
WHERE group_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_student_unique ON users(first_name, last_name, role_id, group_id)
WHERE group_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_users_group_id ON users(group_id);
CREATE INDEX IF NOT EXISTS idx_users_first_last_name ON users(first_name, last_name);
CREATE INDEX IF NOT EXISTS idx_users_role_first_last ON users(role_id, first_name, last_name);
CREATE INDEX IF NOT EXISTS idx_schedule_weekday ON schedule(weekday);
CREATE INDEX IF NOT EXISTS idx_schedule_teacher_id ON schedule(teacher_id);
CREATE INDEX IF NOT EXISTS idx_schedule_group_id ON schedule(group_id);
CREATE INDEX IF NOT EXISTS idx_schedule_subject_id ON schedule(subject_id);
CREATE INDEX IF NOT EXISTS idx_schedule_weekday_start_time ON schedule(weekday, start_time);
CREATE INDEX IF NOT EXISTS idx_grades_student_id ON grades(student_id);
CREATE INDEX IF NOT EXISTS idx_grades_teacher_id ON grades(teacher_id);
CREATE INDEX IF NOT EXISTS idx_grades_subject_id ON grades(subject_id);
CREATE INDEX IF NOT EXISTS idx_grades_schedule_id ON grades(schedule_id);
CREATE INDEX IF NOT EXISTS idx_grades_student_subject ON grades(student_id, subject_id);
CREATE INDEX IF NOT EXISTS idx_attendance_student_id ON attendance(student_id);
CREATE INDEX IF NOT EXISTS idx_attendance_schedule_id ON attendance(schedule_id);
CREATE INDEX IF NOT EXISTS idx_attendance_student_schedule ON attendance(student_id, schedule_id);
CREATE INDEX IF NOT EXISTS idx_subjects_teacher_id ON subjects(teacher_id);
CREATE INDEX IF NOT EXISTS idx_subjects_name ON subjects(subject_name);
INSERT INTO roles (role_name)
VALUES ('super_user'),
    ('admin'),
    ('teacher'),
    ('student') ON CONFLICT (role_name) DO NOTHING;
INSERT INTO users (
        name,
        usermax_id,
        first_name,
        last_name,
        role_id,
        group_id
    )
VALUES (
        'Aslan Alibekov',
        94246618,
        'Aslan',
        'Alibekov',
        (
            SELECT role_id
            FROM roles
            WHERE role_name = 'super_user'
        ),
        NULL
    ) ON CONFLICT (usermax_id) DO
UPDATE
SET first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    name = EXCLUDED.name,
    role_id = EXCLUDED.role_id;
