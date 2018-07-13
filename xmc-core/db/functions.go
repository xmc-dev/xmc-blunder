package db

var (
	xmcPageChildren = map[string]string{
		"201806080015": `
CREATE OR REPLACE FUNCTION xmc_page_children() RETURNS void as $$
DECLARE
	m int;
BEGIN
SELECT MAX(nlevel(path)) into m FROM pages;
UPDATE pages SET parent_id = NULL;
FOR i IN 1..m LOOP
	UPDATE pages p1 SET parent_id = (
		SELECT id FROM pages p2
		WHERE p1.path != p2.path
		AND p2.path = subpath(p1.path, 0, -i)
		LIMIT 1
	)
	WHERE p1.parent_id IS NULL
	AND nlevel(p1.path) >= i
	AND p1.path != '';
END LOOP;
END; $$
LANGUAGE plpgsql;
`,
		"201806100030": `
CREATE OR REPLACE FUNCTION xmc_page_children() RETURNS void as $$
DECLARE
	m int;
	last_event timestamp with time zone;
	last_update timestamp with time zone;
BEGIN
SELECT last_page_event INTO last_event FROM internal;
SELECT last_page_children_update INTO last_update FROM internal;
IF last_event = last_update THEN
	RAISE NOTICE 'nothing to update';
	RETURN;
END IF;
SELECT MAX(nlevel(path)) INTO m FROM pages;
UPDATE pages SET parent_id = NULL;
FOR i IN 1..m LOOP
	UPDATE pages p1 SET parent_id = (
		SELECT id FROM pages p2
		WHERE p1.path != p2.path
		AND p2.path = subpath(p1.path, 0, -i)
		LIMIT 1
	)
	WHERE p1.parent_id IS NULL
	AND nlevel(p1.path) >= i
	AND p1.path != '';
UPDATE internal SET last_page_children_update = last_event;
END LOOP;
END; $$
LANGUAGE plpgsql;
`,
	}
)
