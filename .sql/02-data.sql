\connect db;

WITH tmp AS (
    INSERT INTO board.user (Username, Emailaddress) VALUES
    ('CoolAssMitch420', 'evilmitch@evilmoneydance.com'),
    ('EvilAssMitch666', 'coolassmitch@evilashell.com') 
    RETURNING Id
)
INSERT INTO board.thread (UserId, Title) 
    SELECT Id, 'A Camaro With Two Dragons' from tmp;

UPDATE board.user SET UserPassword = decode(crypt('coolassmitch420', gen_salt('bf', 8)), 'escape') WHERE Username = 'CoolAssMitch420';
UPDATE board.user SET UserPassword = decode(crypt('evilassmitch666', gen_salt('bf', 8)), 'escape') WHERE Username = 'EvilAssMitch666';
