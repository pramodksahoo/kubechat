import '@xterm/xterm/css/xterm.css';

import { MutableRefObject, useEffect, useMemo, useRef, useState } from "react";

import { Button } from '@/components/ui/button';
import { ChevronsDown } from 'lucide-react';
import { FitAddon } from '@xterm/addon-fit';
import { PodSocketResponse } from '@/types';
import { SearchAddon } from '@xterm/addon-search';
import { Terminal } from '@xterm/xterm';
import { clearLogs } from '@/data/Workloads/Pods/PodLogsSlice';
import { getSystemTheme } from '@/utils';
import { useAppDispatch } from '@/redux/hooks';

type XtermProp = {
  containerNameProp: string;
  updateLogs: (currentLog: PodSocketResponse) => void;
  xterm: MutableRefObject<Terminal | null>
  searchAddonRef: MutableRefObject<SearchAddon | null>
};

const DARK_THEME = {
  background: "#0F141B",
  foreground: "#DFE7FF",
  cursor: "#DFE7FF",
  selectionBackground: "#2D394A",
};

const LIGHT_THEME = {
  background: "#FFFFFF",
  foreground: "#0F141B",
  cursor: "#0F141B",
  selectionBackground: "#BCC6DC",
};

const XtermTerminal = ({ containerNameProp, xterm, searchAddonRef, updateLogs }: XtermProp) => {
  const dispatch = useAppDispatch();
  const terminalRef = useRef<HTMLDivElement | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);

  const fitAddon = useRef<FitAddon | null>(null);
  const [showScrollDown, setShowScrollDown] = useState(false);

  useEffect(() => {
    const newContainer = `-------------------${containerNameProp || 'All Containers'}-------------------`;
    xterm?.current?.writeln(newContainer);
    updateLogs({log: newContainer} as PodSocketResponse);
  }, [containerNameProp, updateLogs, xterm]);

  const scrollToBottom = () => {
    const xtermContainer = document.querySelector('.xterm-viewport');
    if (xtermContainer) {
      xtermContainer.scrollTop = xtermContainer.scrollHeight;
    }
  };

  const theme = useMemo(() => (getSystemTheme() === "light" ? LIGHT_THEME : DARK_THEME), []);

  useEffect(() => {
    if (terminalRef.current && xterm) {
      xterm.current = new Terminal({
        cursorBlink: false,
        theme,
        scrollback: 9999999,
        fontSize: 13
      });

      fitAddon.current = new FitAddon();
      searchAddonRef.current = new SearchAddon();
      xterm.current.loadAddon(fitAddon.current);
      xterm.current.loadAddon(searchAddonRef.current);
      xterm.current.open(terminalRef.current);

      // Fit the terminal to the container
      fitAddon.current.fit();

      // Resize the terminal on window resize
      const handleResize = () => fitAddon.current?.fit();
      window.addEventListener('resize', handleResize);

      // Add ResizeObserver to handle container resize (for Resizable component)
      const resizeObserver = new ResizeObserver(() => {
        // Use requestAnimationFrame to ensure resize happens after DOM updates
        requestAnimationFrame(() => {
          fitAddon.current?.fit();
        });
      });
      
      if (containerRef.current) {
        resizeObserver.observe(containerRef.current);
      }

      const xtermContainer = document.querySelector('.xterm-viewport');

      const checkIfBottom = () => {
        const xtermContainer = document.querySelector('.xterm-viewport');
        if (xtermContainer && xtermContainer?.clientHeight + xtermContainer?.scrollTop < xtermContainer.scrollHeight) {
          setShowScrollDown(true);
        } else {
          setShowScrollDown(false);
        }
      };
      xtermContainer?.addEventListener('scroll', checkIfBottom);
      
      return () => {
        xterm.current?.dispose();
        window.removeEventListener('resize', handleResize);
        resizeObserver.disconnect();
        xtermContainer?.removeEventListener('scroll', checkIfBottom);
        dispatch(clearLogs());
      };
    }
  }, [dispatch, searchAddonRef, theme, xterm]);

  return (
    <div ref={containerRef} className="w-full h-full relative">
      {
        showScrollDown &&
        <Button
          variant="secondary"
          size="icon"
          className='absolute bottom-10 right-0 mt-1 mr-2 rounded z-10 border border-border/60'
          onClick={scrollToBottom}
        >  <ChevronsDown className="h-4 w-4" />
        </Button>
      }

      <div ref={terminalRef} />
    </div>
  );
};

export default XtermTerminal;
