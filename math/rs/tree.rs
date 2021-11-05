use std::ptr::NonNull;

pub enum NodeType {
    Terminal(f64),
    NonTerminal(String),
}

pub struct Node {
    pub node_type: NodeType,
    pub children: Option<Vec<NonNull<Node>>>,
}

impl Node {
    pub fn new_terminal(value: f64) -> Self {
        Self {
            node_type: NodeType::Terminal(value),
            children: None,
        }
    }

    pub fn new_non_terminal(operator: String) -> Self {
        Self {
            node_type: NodeType::NonTerminal(operator),
            children: Some(Vec::new()),
        }
    }
}
