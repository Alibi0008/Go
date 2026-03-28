-- Добавляем 20 пользователей
INSERT INTO users (name, email, gender, birth_date) VALUES
                                                        ('Ali','ali@mail.com','male','2000-01-01'),
                                                        ('John','john@mail.com','male','1999-02-02'),
                                                        ('Sara','sara@mail.com','female','2001-03-03'),
                                                        ('Mike','mike@mail.com','male','1998-04-04'),
                                                        ('Anna','anna@mail.com','female','2002-05-05'),
                                                        ('Tom','tom@mail.com','male','1995-06-06'),
                                                        ('Lucy','lucy@mail.com','female','1997-07-07'),
                                                        ('Max','max@mail.com','male','1990-08-08'),
                                                        ('Kate','kate@mail.com','female','1992-09-09'),
                                                        ('Leo','leo@mail.com','male','1993-10-10'),
                                                        ('Nina','nina@mail.com','female','1994-11-11'),
                                                        ('Paul','paul@mail.com','male','1996-12-12'),
                                                        ('Olga','olga@mail.com','female','1998-01-13'),
                                                        ('Mark','mark@mail.com','male','1991-02-14'),
                                                        ('Emma','emma@mail.com','female','1999-03-15'),
                                                        ('Gleb','gleb@mail.com','male','1989-04-16'),
                                                        ('Rita','rita@mail.com','female','1995-05-17'),
                                                        ('Vlad','vlad@mail.com','male','1997-06-18'),
                                                        ('Jane','jane@mail.com','female','1994-07-19'),
                                                        ('Oleg','oleg@mail.com','male','1992-08-20');

-- Добавляем связи (друзей)
-- Делаем так, чтобы у Ali (id=1) и John (id=2) были 3 общих друга: Sara (id=3), Mike (id=4) и Anna (id=5).
INSERT INTO user_friends (user_id, friend_id) VALUES
-- Друзья Ali (id=1)
(1, 3), (1, 4), (1, 5),
-- Друзья John (id=2)
(2, 3), (2, 4), (2, 5);