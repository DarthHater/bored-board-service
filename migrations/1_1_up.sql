CREATE TABLE board.user
(
    Id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    Username varchar(250) UNIQUE,
    Emailaddress varchar(250) UNIQUE,
    UserPassword bytea,
    UserRole int
);

CREATE TABLE board.thread
(
    Id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    UserId UUID,
    Title varchar(250),
    PostedAt TIMESTAMP DEFAULT now(),
    Deleted boolean DEFAULT false
);

CREATE INDEX thread_deleted_idx ON board.thread (Deleted);

CREATE TABLE board.thread_post
(
    Id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ThreadId UUID REFERENCES board.thread (Id),
    UserId UUID,
    Body text,
    PostedAt TIMESTAMP DEFAULT now(),
    Deleted boolean DEFAULT false
);

CREATE INDEX thread_post_deleted_idx ON board.thread_post (Deleted);


WITH tmp AS (
    INSERT INTO board.user (Username, Emailaddress, UserPassword, UserRole) VALUES
    ('CoolAssMitch420', 'evilmitch@evilmoneydance.com', decode(crypt('coolassmitch420', gen_salt('bf', 8)), 'escape'), 0),
    ('EvilAssMitch666', 'coolassmitch@evilashell.com', decode(crypt('evilassmitch666', gen_salt('bf', 8)), 'escape'), 3)
    RETURNING Id
)

INSERT INTO board.thread (UserId, Title)
    SELECT Id, 'A Camaro With Two Dragons' from tmp;

INSERT INTO board.thread_post (ThreadId, UserId, Body) VALUES
    ((SELECT bt.Id FROM board.thread bt INNER JOIN board.user bu on bu.Id = bt.UserId WHERE bu.Username = 'CoolAssMitch420'), (SELECT Id FROM board.user WHERE Username = 'CoolAssMitch420'), 'Test reply'),
    ((SELECT bt.Id FROM board.thread bt INNER JOIN board.user bu on bu.Id = bt.UserId WHERE bu.Username = 'EvilAssMitch666'), (SELECT Id FROM board.user WHERE Username = 'EvilAssMitch666'), 'Test reply');
    
