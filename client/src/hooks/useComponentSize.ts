import { useRef, useState, useLayoutEffect } from "react";

export const useComponentSize = <T extends HTMLElement>() => {
  const ref = useRef<T>(null);
  const [size, setSize] = useState({ width: 0, height: 0 });

  useLayoutEffect(() => {
    if (ref.current) {
      const updateSize = () => {
        setSize({
          width: ref.current!.offsetWidth,
          height: ref.current!.offsetHeight,
        });
      };

      updateSize();

      const resizeObserver = new ResizeObserver(updateSize);
      resizeObserver.observe(ref.current);

      return () => resizeObserver.disconnect();
    }
  }, []);

  return [ref, size] as const;
};
