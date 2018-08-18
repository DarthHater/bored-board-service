CREATE TYPE dup_result AS (f1 int, f2 text);

CREATE FUNCTION dup(int) RETURNS dup_result
    AS $$
    SELECT $1, CAST($1 AS text) || ' is text'
    $$
    LANGUAGE SQL;

SELECT * FROM dup(42);

CREATE TYPE user_info_result AS (threads int, posts int, last_posted text);

CREATE FUNCTION get_user_info(text) RETURNS user_info_result
    AS
    $$
        SELECT COUNT(t)::int, COUNT(tp)::int, t.title::text
            FROM board.thread t
                INNER JOIN board.thread_post tp ON t.id = tp.threadid
                INNER JOIN board.user u on tp.userid = u.id
        WHERE u.id::text = $1
        GROUP BY t.title
    $$
    LANGUAGE SQL;

SELECT * FROM get_user_info('86f98ffc-4214-4e0e-ab9a-228bd98b35e3');

DROP TYPE user_info_result;

SELECT COUNT(t), COUNT(tp), t.title::text
            FROM board.thread t
                INNER JOIN board.thread_post tp ON t.id = tp.threadid
                INNER JOIN board.user u on tp.userid = u.id
        WHERE u.id::text = '86f98ffc-4214-4e0e-ab9a-228bd98b35e3'
        GROUP BY t.title