const stripTrailingSlash = (value) => {
  if (!value || value === "/") {
    return "";
  }

  return value.endsWith("/") ? value.slice(0, -1) : value;
};

export const basePath = stripTrailingSlash(import.meta.env.BASE_URL);

export const withBase = (path) => `${basePath}${path}`;
