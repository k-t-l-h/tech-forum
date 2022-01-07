CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users(
    email citext UNIQUE NOT NULL,
    fullname varchar NOT NULL,
    nickname citext COLLATE "C" UNIQUE PRIMARY KEY,
    about text NOT NULL DEFAULT ''
);
--оставить оба.
CREATE unique INDEX users_nickname ON users(nickname);  --тест
CREATE unique INDEX users_email ON users(email); --ускорили вставку постов
--CREATE INDEX users_full ON users(email, nickname);  --ускорили вставку постов

CREATE UNLOGGED TABLE forums (
    title varchar NOT NULL,
    author citext COLLATE "C" ,
    slug citext PRIMARY KEY,
    posts int DEFAULT 0,
    threads int DEFAULT 0
);
CREATE unique INDEX forums_slug ON forums(slug);
--CREATE INDEX forums_users ON forums(author); --замедлило вставку постов, ускорило всё остальное

CREATE UNLOGGED TABLE forum_users (
    nickname citext  collate "C",
    forum citext COLLATE "C" ,
    CONSTRAINT fk UNIQUE(nickname, forum)
);
--CREATE INDEX fu_nickname ON forum_users USING hash(nickname);
CREATE INDEX fu_forum ON forum_users(forum);
CREATE INDEX fu_full ON forum_users(nickname,forum);

CREATE UNLOGGED TABLE threads (
    id serial PRIMARY KEY,
    author  citext COLLATE "C" ,
    message citext NOT NULL,
    title citext NOT NULL,
    created_at timestamp with time zone,
    forum  citext COLLATE "C" ,
    slug citext,
    votes int
);

--CREATE INDEX IF NOT EXISTS threads_slug ON threads USING hash(slug); --тест
--CREATE INDEX IF NOT EXISTS threads_id ON threads USING hash(id); --тест
CREATE INDEX IF NOT EXISTS threads_forum ON threads(forum); --не убирать
CREATE INDEX IF NOT EXISTS created_forum_index ON threads(forum, created_at);
CREATE INDEX  IF NOT EXISTS cluster_thread ON threads(id, forum); --ускоряет
CREATE INDEX ON threads(slug, id, forum);

CREATE UNLOGGED TABLE posts (
    id serial  PRIMARY KEY ,
    author citext COLLATE "C",
    post text NOT NULL,
    created_at timestamp with time zone,
    forum citext COLLATE "C",
    isEdited bool,
    parent int,
    thread int,
    path integer []
);
CREATE INDEX IF NOT EXISTS posts_thread ON posts(thread); --не убирать
CREATE INDEX IF NOT EXISTS posts_parent_thread_index ON posts(parent, thread);
CREATE INDEX parent_tree_index ON posts ((path[1]), path, id);
/*
CREATE INDEX IF NOT EXISTS posts_id ON posts thread, created_at, id, parent, path);
CREATE INDEX pdesc ON posts(thread, path);
CREATE INDEX ptida ON posts(thread, id);

CREATE INDEX parent_tree_index4 ON posts (id, (path[1]));
 */


CREATE UNLOGGED TABLE votes (
    author citext COLLATE "C",
    vote int,
    thread int,
    CONSTRAINT checks UNIQUE(author, thread)
);
CREATE INDEX votes_full ON votes(author, vote, thread);

CREATE OR REPLACE FUNCTION update_path() RETURNS TRIGGER AS
$update_path$
DECLARE
parent_path integer[];
    parent_thread int;
BEGIN
SELECT path FROM posts WHERE id = new.parent INTO parent_path;
NEW.path := parent_path || new.id;
RETURN new;
END
$update_path$ LANGUAGE plpgsql;


CREATE TRIGGER path_update_trigger
    BEFORE INSERT
    ON posts
    FOR EACH ROW
    EXECUTE PROCEDURE update_path();

VACUUM;
VACUUM ANALYSE;