CREATE TABLE contents (
    id INT NOT NULL AUTO_INCREMENT,
    title TEXT,
    body TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

INSERT INTO contents (id,title, body) VALUES (1,'吾輩は猫である' ,'吾輩は猫である。名前はまだ無い。');
INSERT INTO contents (id,title, body) VALUES (2,'人間失格' ,'私は、その男の写真を三葉、見たことがある。');