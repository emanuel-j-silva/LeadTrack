-- 000002_seed.up.sql
INSERT INTO question (label, weight) VALUES
('Entrega de Resultados', 25),
('Execução e Qualidade do Trabalho', 20),
('Capacidade de Aprendizado e Desenvolvimento', 20),
('Resolução de Problemas e Pensamento Crítico', 15),
('Colaboração, Influência e Liderança', 10),
('Visão Estratégica e Potencial de Crescimento', 10);

INSERT INTO employee (id, name, email, position_name) VALUES
(1, 'Alice', 'alice@company.com', 'CEO'),
(2, 'Bob', 'bob@company.com', 'CTO'),
(3, 'Carol', 'carol@company.com', 'CFO'),
(4, 'David', 'david@company.com', 'Eng Manager'),
(5, 'Eva', 'eva@company.com', 'Eng Manager'),
(6, 'Frank', 'frank@company.com', 'PM'),
(7, 'Grace', 'grace@company.com', 'Dev'),
(8, 'Henry', 'henry@company.com', 'Sr Engineer'),
(9, 'Isabelle', 'isabelle@company.com', 'Dev'),
(10, 'James', 'james@company.com', 'Dev'),
(11, 'Karen', 'karen@company.com', 'Dev'),
(12, 'Liam', 'liam@company.com', 'Dev'),
(13, 'Mia', 'mia@company.com', 'Dev'),
(14, 'Noah', 'noah@company.com', 'Dev'),
(15, 'Olivia', 'olivia@company.com', 'Dev'),
(16, 'Paul', 'paul@company.com', 'Dev'),
(17, 'Quinn', 'quinn@company.com', 'Dev'),
(18, 'Rachel', 'rachel@company.com', 'Analyst'),
(19, 'Samuel', 'samuel@company.com', 'Analyst'),
(20, 'Tina', 'tina@company.com', 'HR');

SELECT SETVAL('employee_id_seq', (SELECT MAX(id) FROM employee));

INSERT INTO leader_lead (leader_id, lead_id) VALUES
(1, 2), (1, 3), (1, 6), (1, 20),
(2, 4), (2, 5), (2, 7), (2, 17), (2, 16),
(4, 8), (4, 12),
(8, 10), (8, 11),
(5, 9), (5, 13), (5, 14),
(3, 18), (3, 19),
(6, 15);
