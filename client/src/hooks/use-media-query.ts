import { useEffect, useState } from "react";

const getInitialValue = (query: string) => {
  if (typeof window === "undefined") {
    return false;
  }
  return window.matchMedia(query).matches;
};

export const useMediaQuery = (query: string) => {
  const [matches, setMatches] = useState(() => getInitialValue(query));

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    const mediaQueryList = window.matchMedia(query);
    const handler = (event: MediaQueryListEvent) => setMatches(event.matches);

    mediaQueryList.addEventListener("change", handler);
    setMatches(mediaQueryList.matches);

    return () => mediaQueryList.removeEventListener("change", handler);
  }, [query]);

  return matches;
};
