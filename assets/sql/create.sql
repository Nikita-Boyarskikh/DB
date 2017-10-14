CREATE TABLE IF NOT EXISTS users (
  nickname CITEXT, --CONSTRAINT pk__users_ID PRIMARY KEY,
  --CONSTRAINT ch__users_nickname CHECK (nickname ~ '^[\d\w_]+$'),
  fullname VARCHAR,
  --CONSTRAINT nn__users_fullname NOT NULL,
  email    CITEXT,
  --CONSTRAINT nn__users_email NOT NULL
  --CONSTRAINT uk__users_email UNIQUE
  --CONSTRAINT ch__users_email CHECK (email ~* '^[\w\d._-]+@[\w\d._-]+\.\w{2,4}$'),
  about    VARCHAR
);

CREATE TABLE IF NOT EXISTS forums (
  ID      SERIAL, --CONSTRAINT pk__forums_ID PRIMARY KEY,
  posts   INT DEFAULT 0,
  slug    CITEXT,
  --CONSTRAINT nn__forums_slug NOT NULL
  --CONSTRAINT uk__forums_slug UNIQUE,
  threads INT DEFAULT 0,
  title   VARCHAR,
  --CONSTRAINT nn__forums_title NOT NULL,
  userID  CITEXT
  --CONSTRAINT nn__forums_userID NOT NULL
  --CONSTRAINT uk__forums_userID UNIQUE
  --CONSTRAINT fk__forums_user__users_ID REFERENCES users (Nickname)
  --ON UPDATE RESTRICT ON DELETE RESTRICT
  --CONSTRAINT ch__forums_slug CHECK (slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$')
);

CREATE TABLE IF NOT EXISTS threads (
  ID       SERIAL, --CONSTRAINT pk__threads_ID PRIMARY KEY,
  authorID CITEXT,
  --CONSTRAINT nn__threads_authorID NOT NULL
  --CONSTRAINT fk__threads_authorID__user_ID REFERENCES users (Nickname)
  --ON UPDATE RESTRICT ON DELETE RESTRICT,
  created  TIMESTAMPTZ(3) DEFAULT now(),
  forumID  CITEXT,
  --CONSTRAINT nn__threads_forumID NOT NULL
  --CONSTRAINT fk__threads_forumID__forums_ID REFERENCES forums (ID)
  --ON UPDATE RESTRICT ON DELETE RESTRICT,
  message  TEXT,
  --CONSTRAINT nn__threads_message NOT NULL,
  title    VARCHAR,
  --CONSTRAINT nn__threads_title NOT NULL,
  slug     CITEXT,
  --CONSTRAINT uk__threads_slug UNIQUE,
  votes    INT            DEFAULT 0
  --CONSTRAINT ch__threads_slug CHECK (slug ~ '^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$')
);

CREATE TABLE IF NOT EXISTS posts (
  ID       SERIAL, --CONSTRAINT pk__posts_ID PRIMARY KEY,
  authorID CITEXT,
  --CONSTRAINT nn__posts_authorID NOT NULL
  --CONSTRAINT fk__posts_authorID__users_ID REFERENCES users (Nickname)
  --ON UPDATE RESTRICT ON DELETE RESTRICT,
  created  TIMESTAMP(3) DEFAULT now(),
  forumID  CITEXT,
  --CONSTRAINT nn__posts_forumID NOT NULL
  --CONSTRAINT fk__posts_forumID__forums_ID REFERENCES forums (ID)
  --ON UPDATE RESTRICT ON DELETE RESTRICT,
  isEdited BOOLEAN      DEFAULT FALSE,
  --CONSTRAINT nn__posts_isEdited NOT NULL,
  message  TEXT,
  --CONSTRAINT nn__posts_message NOT NULL,
  parentID INT DEFAULT 0,
  --CONSTRAINT fk__posts_parentID__posts_ID REFERENCES posts (ID),
  threadID INT
  --CONSTRAINT nn__posts_threadID NOT NULL
  --CONSTRAINT fk__posts_threadID__threads_ID REFERENCES threads (ID)
  --ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS voices (
  threadID INT,
  --CONSTRAINT nn__voices_threadID NOT NULL
  --CONSTRAINT fk__voices_threadID__threads_ID REFERENCES threads (ID)
  --ON UPDATE RESTRICT ON DELETE RESTRICT
  userID   CITEXT,
  --CONSTRAINT nn__voices_userID NOT NULL
  --CONSTRAINT fk__voices_userID__users_ID REFERENCES users (ID)
  --ON UPDATE RESTRICT ON DELETE RESTRICT,
  voice    INT,


  --CONSTRAINT ch__voices_voice CHECK (voice IN (-1, 1)),
  CONSTRAINT pk__voices_thread_nickname PRIMARY KEY (threadID, userID)
);

CREATE INDEX IF NOT EXISTS idx__users_nickname
  ON users (nickname);
CREATE INDEX IF NOT EXISTS idx__users_email
  ON users (email);

CREATE INDEX IF NOT EXISTS idx__threads_ID
  ON threads (ID);
CREATE INDEX IF NOT EXISTS idx__threads_slug
  ON threads (slug)
  WHERE slug IS NOT NULL;