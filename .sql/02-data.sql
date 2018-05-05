\connect db;

WITH tmp AS (
    INSERT INTO board.user (Username, Emailaddress) VALUES
    ('CoolAssMitch420', 'evilmitch@evilmoneydance.com'),
    ('EvilAssMitch666', 'coolassmitch@evilashell.com') 
    RETURNING Id
)

INSERT INTO board.thread (UserId, Title) 
    SELECT Id, 'A Camaro With Two Dragons' from tmp;

INSERT INTO board.user_roles (UserId, RoleId)
    SELECT UserId, 0 from tmp;
