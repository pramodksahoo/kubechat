import { memo } from "react";

type CurrentByDesiredCellProps = {
  cellValue: string;
};


const CurrentByDesiredCell = memo(function ({ cellValue }: CurrentByDesiredCellProps) {
  const valueArray = cellValue.split('/');
  const isReady = valueArray[0] === valueArray[1];
  return (
    <span className={`text-sm truncate px-3 ${isReady ? 'text-success' : 'text-destructive'}`}>
      {cellValue}
    </span>
  );
});

export {
  CurrentByDesiredCell
};
