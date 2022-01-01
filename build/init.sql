CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users(
    email citext UNIQUE NOT NULL,
    fullname varchar NOT NULL,
    nickname citext COLLATE ucs_basic UNIQUE PRIMARY KEY,
    about text NOT NULL DEFAULT ''
);

CREATE UNLOGGED TABLE forums (
    title varchar NOT NULL,
    author citext references users(nickname),
    slug citext PRIMARY KEY,
    posts int DEFAULT 0,
    threads int DEFAULT 0
);

CREATE UNLOGGED TABLE forum_users (
    nickname citext references users(nickname),
    forum citext references forums(slug),
    CONSTRAINT fk UNIQUE(nickname, forum)
);

CREATE INDEX fu_nick ON forum_users(nickname);
CREATE INDEX fu_for ON forum_users(forum);

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
CREATE INDEX ON threads(id, forum);
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

CREATE INDEX parent_tree_index
    ON posts ((path[1]), path DESC, id);

CREATE INDEX parent_tree_index2
    ON posts (id, (path[1]));
