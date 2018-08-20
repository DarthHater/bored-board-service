DROP INDEX thread_post_deleted_idx;
DROP TABLE IF EXISTS board.thread_post;

DROP INDEX thread_deleted_idx;
DROP TABLE IF EXISTS board.thread;

DROP TABLE IF EXISTS "board"."user";
