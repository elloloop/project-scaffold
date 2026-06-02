pub mod platform {
    pub fn display_name(value: &str) -> String {
        value
            .split(['-', '_', ' '])
            .filter(|part| !part.is_empty())
            .map(|part| {
                let mut chars = part.chars();
                match chars.next() {
                    Some(first) => first.to_uppercase().collect::<String>() + chars.as_str(),
                    None => String::new(),
                }
            })
            .collect::<Vec<_>>()
            .join(" ")
    }
}

pub mod serverkit {
    use crate::platform;

    #[derive(Debug, Eq, PartialEq)]
    pub struct HealthResponse {
        pub service: String,
        pub status: &'static str,
    }

    pub fn health(service: &str) -> HealthResponse {
        HealthResponse {
            service: platform::display_name(service),
            status: "ok",
        }
    }
}

pub mod testkit {
    pub fn assert_equal<T>(got: T, want: T)
    where
        T: std::fmt::Debug + PartialEq,
    {
        assert_eq!(got, want);
    }
}

#[cfg(test)]
mod tests {
    use super::{serverkit, testkit};

    #[test]
    fn creates_health_response() {
        let response = serverkit::health("project_scaffold");

        testkit::assert_equal(response.service, "Project Scaffold".to_string());
        testkit::assert_equal(response.status, "ok");
    }
}
