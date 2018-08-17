-- \connect db;

-- WITH tmp AS (
--     INSERT INTO board.user (Username, Emailaddress, UserPassword, UserRole) VALUES
--     ('CoolAssMitch420', 'evilmitch@evilmoneydance.com', decode(crypt('coolassmitch420', gen_salt('bf', 8)), 'escape'), 0),
--     ('EvilAssMitch666', 'coolassmitch@evilashell.com', decode(crypt('evilassmitch666', gen_salt('bf', 8)), 'escape'), 3)
--     RETURNING Id
-- ),
-- thread AS (
-- 	INSERT INTO board.thread (UserId, Title)
-- 	SELECT Id, 'A Camaro With Two Dragons' from tmp
-- 	RETURNING Id, UserId
-- )

-- INSERT INTO board.thread (UserId, Title)
--     SELECT Id, 'A Camaro With Two Dragons' from tmp;

-- INSERT INTO board.thread_post (ThreadId, UserId, Body) VALUES
--     ((SELECT bt.Id FROM board.thread bt INNER JOIN board.user bu on bu.Id = bt.UserId WHERE bu.Username = 'CoolAssMitch420'), (SELECT Id FROM board.user WHERE Username = 'CoolAssMitch420'), 'Test reply'),
--     ((SELECT bt.Id FROM board.thread bt INNER JOIN board.user bu on bu.Id = bt.UserId WHERE bu.Username = 'EvilAssMitch666'), (SELECT Id FROM board.user WHERE Username = 'EvilAssMitch666'), 'Test reply');
