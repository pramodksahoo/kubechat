import { useEffect, useRef, useState } from "react";

import { getDisplayTime } from "@/utils";

type TimeCellProps = {
  cellValue: string;
};


function TimeCell({ cellValue }: TimeCellProps) {
  const [currentTime, setCurrentTime] = useState(() => new Date().getTime() - new Date(cellValue).getTime());
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
    }
    timerRef.current = setInterval(() => {
      setCurrentTime((previous) => previous + 500);
    }, 1000);

    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
    };
  }, []);

  useEffect(() => {
    setCurrentTime(new Date().getTime() - new Date(cellValue).getTime());
  }, [cellValue]);
  return (
    <div className="px-3">
      <span title={cellValue} className="text-sm text-gray-700 dark:text-gray-100">
        {getDisplayTime(Number(currentTime))}
      </span>
    </div>
  );
}

export {
  TimeCell
};
