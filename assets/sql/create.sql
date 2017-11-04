SET WORK_MEM = '1GB';

CREATE TABLE IF NOT EXISTS users (
  nickname CITEXT CONSTRAINT pk__users_nickname PRIMARY KEY,
  fullname TEXT,
  email    CITEXT,
  about    TEXT
);

CREATE TABLE IF NOT EXISTS forums (
  ID      SERIAL CONSTRAINT pk__forums_ID PRIMARY KEY,
  posts   INT8 DEFAULT 0,
  slug    CITEXT,
  threads INT4 DEFAULT 0,
  title   TEXT,
  userID  CITEXT
);

CREATE TABLE IF NOT EXISTS threads (
  ID       SERIAL4 CONSTRAINT pk__threads_ID PRIMARY KEY,
  authorID CITEXT,
  created  TIMESTAMPTZ(3) DEFAULT now(),
  forumID  CITEXT,
  message  TEXT,
  title    TEXT,
  slug     CITEXT,
  votes    INT4           DEFAULT 0
);

CREATE TABLE IF NOT EXISTS posts (
  ID       SERIAL8 CONSTRAINT pk__posts_ID PRIMARY KEY,
  authorID CITEXT,
  created  TIMESTAMP(3) DEFAULT now(),
  forumID  CITEXT,
  isEdited BOOLEAN      DEFAULT FALSE,
  message  TEXT,
  parentID INT8         DEFAULT 0,
  threadID INT4,
  parents  INT8 []
);

CREATE TABLE IF NOT EXISTS voices (
  threadID INT4,
  userID   CITEXT,
  voice    INT2,
  CONSTRAINT pk__voices_threadID_userID PRIMARY KEY (threadID, userID)
);

CREATE INDEX IF NOT EXISTS idx__users_email
  ON users (email);

CREATE INDEX IF NOT EXISTS idx__threads_created
  ON threads (created);
CREATE INDEX IF NOT EXISTS idx__threads_slug
  ON threads (slug)
  WHERE slug IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx__threads_forumID
  ON threads (forumID);

CREATE INDEX IF NOT EXISTS idx__forums_slug
  ON forums (slug);

CREATE INDEX IF NOT EXISTS idx__posts_forumID
  ON posts (forumID);
CREATE INDEX IF NOT EXISTS idx__posts_authorID
  ON posts (authorID);
CREATE INDEX IF NOT EXISTS idx__posts_threadID
  ON posts (threadID);
CREATE INDEX IF NOT EXISTS idx__posts_parentID
  ON posts (parentID);
CREATE INDEX IF NOT EXISTS idx__posts_created
  ON posts (created);