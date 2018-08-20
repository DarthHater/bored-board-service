CREATE TABLE board.message
(
    Id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    UserId UUID,
    Title varchar(250),
    PostedAt TIMESTAMP DEFAULT now(),
    Deleted boolean DEFAULT false
);

CREATE TABLE board.message_post
(
    Id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    MessageId UUID REFERENCES board.message (Id),
    UserId UUID,
    Body text,
    PostedAt TIMESTAMP DEFAULT now(),
    Deleted boolean DEFAULT false
);

CREATE TABLE board.message_member
(
    UserId UUID REFERENCES board.user (Id),
    MessageId UUID REFERENCES board.message (Id),
    PostedAt TIMESTAMP DEFAULT now(),
    Deleted boolean NOT NULL DEFAULT false
);
