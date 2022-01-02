CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users(
    email citext UNIQUE NOT NULL,
    fullname varchar NOT NULL,
    nickname citext COLLATE ucs_basic UNIQUE PRIMARY KEY,
    about text NOT NULL DEFAULT ''
);
CREATE INDEX users_email ON users(nickname, email);

CREATE UNLOGGED TABLE forums (
    title varchar NOT NULL,
    author citext references users(nickname),
    slug citext PRIMARY KEY,
    posts int DEFAULT 0,
    threads int DEFAULT 0
);

CREATE INDEX forums_users ON forums(author); --замедлило вставку постов
CREATE UNLOGGED TABLE forum_users (
    nickname citext references users(nickname),
    forum citext references forums(slug),
    CONSTRAINT fk UNIQUE(nickname, forum)
);

CREATE INDEX fu_nick ON forum_users(nickname,forum);

CREATE UNLOGGED TABLE threads (
    id serial PRIMARY KEY,
    author citext references users(nickname),
    message citext NOT NULL,
    title citext NOT NULL,
    created_at timestamp with time zone,
    forum citext references forums(slug),
    slug citext,
    votes int
);

CREATE INDEX IF NOT EXISTS threads_forum ON threads(forum);
CREATE INDEX IF NOT EXISTS created_forum_index ON threads(forum, created_at);
CREATE INDEX ON threads(id, forum); --ускоряет
CREATE INDEX ON threads(slug, id, forum);

CREATE UNLOGGED TABLE posts (
    id serial PRIMARY KEY ,
    author citext references users(nickname),
    post citext NOT NULL,
    created_at timestamp with time zone,
    forum citext references forums(slug),
    isEdited bool,
    parent int,
    thread int references threads(id),
    path  INTEGER[]
);


CREATE INDEX pdesc ON posts(thread, path DESC);
CREATE INDEX pdesc ON posts(thread, path ASC);
CREATE INDEX IF NOT EXISTS posts_parent_thread_index ON posts(parent, thread);
CREATE INDEX ptidd ON posts(thread, id DESC);
CREATE INDEX ptida ON posts(thread, id ASC);

CREATE UNLOGGED TABLE votes (
    author citext references users(nickname),
    vote int,
    thread int references threads(id),
    CONSTRAINT checks UNIQUE(author, thread)
);


CREATE OR REPLACE FUNCTION update_path() RETURNS TRIGGER AS
$update_path$
DECLARE
parent_path  INTEGER[];
    parent_thread int;
BEGIN
    IF (NEW.parent = 0) THEN
        NEW.path := array_append(new.path, new.id);
ELSE
SELECT thread FROM posts WHERE id = new.parent INTO parent_thread;
IF NOT FOUND OR parent_thread != NEW.thread THEN
            RAISE EXCEPTION 'this is an exception' USING ERRCODE = '22000';
end if;

SELECT path FROM posts WHERE id = new.parent INTO parent_path;
NEW.path := parent_path || new.id;
END IF;
RETURN new;
END
$update_path$ LANGUAGE plpgsql;


CREATE TRIGGER path_update_trigger
    BEFORE INSERT
    ON posts
    FOR EACH ROW
    EXECUTE PROCEDURE update_path();
