CREATE OR REPLACE FUNCTION notify_new_article() RETURNS TRIGGER AS $$

    DECLARE 
        article json;
    BEGIN
    
        IF (TG_OP != 'INSERT') THEN
            RETURN NULL;
        END IF;
        
        article = row_to_json(NEW);
        PERFORM pg_notify('new_articles',article::text);
        
        RETURN NULL; 
    END;
    
$$ LANGUAGE plpgsql;

CREATE TABLE articles (
  id BIGSERIAL PRIMARY KEY,
  title TEXT,
  body TEXT,
  category TEXT,
  published_at TIMESTAMP WITH TIME ZONE
);

CREATE TRIGGER articles_notify_on_insert
AFTER INSERT ON articles
    FOR EACH ROW EXECUTE FUNCTION notify_new_article();