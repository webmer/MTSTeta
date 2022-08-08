-- Table: public.tasks
-- DROP TABLE public.tasks;
CREATE TABLE IF NOT EXISTS tasks
(
    uuid uuid PRIMARY KEY NOT NULL,
    name character(50),
    text character(100),
    login character(50),
    status character(20) NOT NULL DEFAULT 'created'::bpchar
);

-- Table: public.approvals
-- DROP TABLE public.approvals;
CREATE TABLE IF NOT EXISTS approvals
(
    id serial PRIMARY KEY NOT NULL,
    task_uuid uuid NOT NULL,
    approval_login character(50) NOT NULL,
    approved boolean,
    sent boolean,
    n integer,
    CONSTRAINT task_uuid FOREIGN KEY (task_uuid)
        REFERENCES public.tasks (uuid) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
        NOT VALID
);

-- Table: public.outbox
-- DROP TABLE public.outbox;
CREATE TABLE IF NOT EXISTS outbox
(
    id SERIAL PRIMARY KEY,
    task_uuid uuid NOT NULL,
    action_timestamp integer NOT NULL,
    type character(10) NOT NULL,
    value character(50) NOT NULL,
    sent boolean,
    CONSTRAINT task_uuid FOREIGN KEY (task_uuid)
        REFERENCES public.tasks (uuid) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
        NOT VALID
);

-- Table: public.outbox_email
-- DROP TABLE public.outbox_email;
CREATE TABLE IF NOT EXISTS outbox_email
(
    id SERIAL PRIMARY KEY,
    task_uuid uuid NOT NULL,
    reciever character(50) NOT NULL,
    type character(10) NOT NULL,
    sent boolean,
    CONSTRAINT task_uuid FOREIGN KEY (task_uuid)
        REFERENCES public.tasks (uuid) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
        NOT VALID
);

INSERT INTO tasks (uuid, name, text, login) VALUES
('437bcb56-0249-479a-b67b-7c4a56a956d8', 'test1', 'this is test task1', 'test123'),
('2281a27e-0ab2-4589-8b06-c4fd5dc6cd45', 'test2', 'this is test task2', 'test123');

INSERT INTO approvals (task_uuid, approval_login, n) VALUES
('437bcb56-0249-479a-b67b-7c4a56a956d8', 'ivan', 1),
('437bcb56-0249-479a-b67b-7c4a56a956d8', 'petr', 2),
('2281a27e-0ab2-4589-8b06-c4fd5dc6cd45', 'test626', 1),
('2281a27e-0ab2-4589-8b06-c4fd5dc6cd45', 'zxvdg', 2);

INSERT INTO outbox (task_uuid, action_timestamp, type, value) VALUES
('437bcb56-0249-479a-b67b-7c4a56a956d8', 1658691169, 'run', 'test123'),
('437bcb56-0249-479a-b67b-7c4a56a956d8', 1658691269, 'approve', 'true'),
('2281a27e-0ab2-4589-8b06-c4fd5dc6cd45', 1658691369, 'run', 'test123'),
('2281a27e-0ab2-4589-8b06-c4fd5dc6cd45', 1658691469, 'approve', 'false');

INSERT INTO outbox_email (task_uuid, reciever, type) VALUES
('437bcb56-0249-479a-b67b-7c4a56a956d8', 'ivan', 'approve'),
('437bcb56-0249-479a-b67b-7c4a56a956d8', 'petr', 'approve'),
('437bcb56-0249-479a-b67b-7c4a56a956d8', 'test123', 'completed'),
('437bcb56-0249-479a-b67b-7c4a56a956d8', 'ivan', 'completed'),
('437bcb56-0249-479a-b67b-7c4a56a956d8', 'petr', 'completed'),
('2281a27e-0ab2-4589-8b06-c4fd5dc6cd45', 'test626', 'approve');
