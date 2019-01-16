ALTER TABLE board.user
ADD COLUMN ConfirmCode varchar(250),
ADD COLUMN Active boolean DEFAULT false;
