CREATE TABLE "user"
(
    id       SERIAL PRIMARY KEY,
    name     VARCHAR(255) NOT NULL,
    email    VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    is_admin BOOLEAN      NOT NULL
);

CREATE TABLE tags
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);


CREATE TABLE features
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);


CREATE TABLE banners
(
    id         SERIAL PRIMARY KEY,
    title      VARCHAR(255) UNIQUE      NOT NULL,
    text       TEXT                     NOT NULL,
    url        VARCHAR(255)             NOT NULL,
    is_active  BOOLEAN                  NOT NULL DEFAULT false,
    feature_id INTEGER                  NOT NULL REFERENCES features (id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);


CREATE TABLE banner_tags
(
    banner_id INTEGER REFERENCES banners (id) NOT NULL,
    tag_id    INTEGER REFERENCES tags (id)    NOT NULL,
    PRIMARY KEY (banner_id, tag_id)
);
