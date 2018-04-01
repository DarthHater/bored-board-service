\connect db;

WITH tmp AS (
    INSERT INTO board.user (Username, Emailaddress) VALUES
    ('CoolAssMitch420', 'evilmitch@evilmoneydance.com'),
    ('EvilAssMitch666', 'coolassmitch@evilashell.com') 
    RETURNING Id
),
thread AS (
	INSERT INTO board.thread (UserId, Title) 
	SELECT Id, 'A Camaro With Two Dragons' from tmp
	RETURNING Id, UserId
)

INSERT INTO board.thread_post (ThreadId, UserId, Body, PostedAt, EditedAt) 
SELECT Id, UserId, 'Som Stuff', now(), now() from thread
