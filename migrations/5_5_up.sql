ALTER TABLE "board"."thread"
ADD COLUMN LastPostedAt TIMESTAMP DEFAULT now();
