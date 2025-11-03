import { Link } from "@tanstack/react-router";
import { memo } from "react";

type NameCellProps = {
  cellValue: string;
  link: string;
};


const NameCell = memo(function ({ cellValue, link}: NameCellProps) {

  return (
    <div className="flex py-0.5">
      <Link
        to={`/${link}`}
      >
        <span title={cellValue} className="max-w-[750px] px-3 text-sm truncate text-primary hover:text-primary/80 hover:underline">
          {cellValue}
        </span>
      </Link>
    </div>

  );
});

export {
  NameCell
};
