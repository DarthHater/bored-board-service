CREATE TABLE board.thread_member
(
    UserId UUID  NOT NULL REFERENCES board.user (Id),
    ThreadId UUID  NOT NULL REFERENCES board.thread (Id),
    LastViewedPostUnixTime int NOT NULL DEFAULT 0,
    PRIMARY KEY (UserId, ThreadId)
);
