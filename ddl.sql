CREATE DATABASE explicai;

CREATE TABLE summaries (
id SERIAL PRIMARY KEY,
external_id UUID NOT NULL,
created_at TIMESTAMP NOT NULL,
updated_at TIMESTAMP NOT NULL,
status VARCHAR(50) NOT NULL,
title VARCHAR(255),
description TEXT,
brief_resume TEXT,
medium_resume TEXT,
progress int,
fulltext TEXT
);

CREATE INDEX idx_external_id ON summaries(external_id);