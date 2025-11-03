import { memo } from "react";

type DefaultCellProps = {
  cellValue: string;
  truncate?: boolean;
};


const DefaultCell = memo(function ({ cellValue, truncate = true }: DefaultCellProps) {
  return (
    <div className="flex">
      <span title={cellValue} className={`max-w-[750px] px-3 text-sm text-foreground/80 ${truncate && 'truncate'}`}>
        {cellValue}
      </span>
    </div>
  );
});

export {
  DefaultCell
};
