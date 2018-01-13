\connect db;

WITH tmp AS (
    INSERT INTO board.user (Username, Emailaddress, UserPassword) VALUES
    ('CoolAssMitch420', 'evilmitch@evilmoneydance.com', decode(crypt('coolassmitch420', gen_salt('bf', 8)), 'escape')),
    ('EvilAssMitch666', 'coolassmitch@evilashell.com', decode(crypt('evilassmitch666', gen_salt('bf', 8)), 'escape')) 
    RETURNING Id
)
INSERT INTO board.thread (UserId, Title) 
    SELECT Id, 'A Camaro With Two Dragons' from tmp;
