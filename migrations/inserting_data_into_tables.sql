INSERT INTO tags (name)
VALUES ('Tag 1'),
       ('Tag 2'),
       ('Tag 3'),
       ('Tag 4'),
       ('Tag 5');


INSERT INTO features (name)
VALUES ('Feature 1'),
       ('Feature 2'),
       ('Feature 3'),
       ('Feature 4'),
       ('Feature 5');


INSERT INTO banners (title, text, url, is_active, feature_id)
VALUES ('Banner 1', 'This is the text of Banner 1', 'https://example.com/banner1', true, 1),
       ('Banner 2', 'This is the text of Banner 2', 'https://example.com/banner2', false, 1),
       ('Banner 3', 'This is the text of Banner 3', 'https://example.com/banner3', true, 2),
       ('Banner 4', 'This is the text of Banner 4', 'https://example.com/banner4', false, 2),
       ('Banner 5', 'This is the text of Banner 5', 'https://example.com/banner5', true, 3),
       ('Banner 6', 'This is the text of Banner 6', 'https://example.com/banner6', false, 3),
       ('Banner 7', 'This is the text of Banner 7', 'https://example.com/banner7', true, 4),
       ('Banner 8', 'This is the text of Banner 8', 'https://example.com/banner8', false, 4),
       ('Banner 9', 'This is the text of Banner 9', 'https://example.com/banner9', true, 5),
       ('Banner 10', 'This is the text of Banner 10', 'https://example.com/banner10', false, 5);


INSERT INTO banner_tags (banner_id, tag_id)
VALUES (1, 1),
       (2, 2),
       (3, 3),
       (4, 4),
       (5, 5),
       (6, 1),
       (7, 2),
       (8, 3),
       (9, 4),
       (10, 5);
