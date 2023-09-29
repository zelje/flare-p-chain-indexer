CREATE USER 'flareindexer'@'%' IDENTIFIED BY 'flareindexerpass';
CREATE DATABASE IF NOT EXISTS flareindexerdb;
GRANT ALL PRIVILEGES ON flareindexerdb.* TO 'flareindexer'@'%';
