ALTER TABLE urls
    ADD CONSTRAINT urls_original_url_unique UNIQUE (original_url);
