import './index.css';

import { HistoryIcon, Maximize2, Minimize2, SettingsIcon, Sparkles, SquarePen, X } from "lucide-react";
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { kcAIStoredChatHistory, kcAIStoredModels } from '@/types/kcAI/addConfiguration';
import { kcDetails, kcList } from '@/routes';
import { useAppDispatch, useAppSelector } from '@/redux/hooks';
import { useEffect, useState } from 'react';

import { Button } from "@/components/ui/button";
import { ChatHistory } from './History';
import { ChatWindow } from '@/components/app/kcAI/Chat';
import { Configuration } from './Configuration';
import { TabsContent } from '@radix-ui/react-tabs';
import { cn } from '@/lib/utils';
import { fetchKcAiTools } from '@/data/KcAi/KcAiToolsSlice';
import { useSidebarSize } from '@/hooks/use-get-sidebar-size';

interface AiChatProps {
  isFullscreen?: boolean
  onToggleFullscreen?: () => void
  customHeight: string
  isDetailsPage?: boolean
  onClose?: () => void
}

export function AiChat({ isFullscreen = false, onToggleFullscreen, customHeight, onClose, isDetailsPage = false }: AiChatProps) {
  const [activeView, setActiveView] = useState("chat");
  const kcAiChatWindow = useSidebarSize("kcai-chat");
  const [kcAIStoredModelsCollection, setKcAIStoredModelsCollection] = useState<kcAIStoredModels>({} as kcAIStoredModels);
  const dispatch = useAppDispatch();
  let config = '';
  let cluster = '';
  if (!isDetailsPage) {
    config = kcList.useParams().config;
    cluster = kcList.useSearch().cluster;
  } else {
    config = kcDetails.useParams().config;
    cluster = kcDetails.useSearch().cluster;
  }
  const {
    clusters,
  } = useAppSelector((state) => state.clusters);
  const clusterConfigKey = `cluster=${cluster}&config=${config}`;
  const getLatestChat = () => {
    const kcAIStoredChatHistory = JSON.parse(localStorage.getItem('kcAIStoredChatHistory') || '{}') as kcAIStoredChatHistory;
    if (kcAIStoredChatHistory && kcAIStoredChatHistory[clusterConfigKey]) {
      const lastKey = Object.keys(kcAIStoredChatHistory[clusterConfigKey]).at(-1);
      if (lastKey) {
        return lastKey;
      }
      return new Date().getTime().toString();
    }
    return new Date().getTime().toString();
  };
  const [currentChatKey, setCurrentChatKey] = useState<string>(getLatestChat);

  useEffect(() => {
    dispatch(fetchKcAiTools({isDev: clusters.version === 'dev', config, cluster}));
    const kcAIStoredModels = JSON.parse(localStorage.getItem('kcAIStoredModels') || '{}') as kcAIStoredModels;
    setKcAIStoredModelsCollection(() => kcAIStoredModels);
  }, []);


  const containerClass = `${customHeight} flex flex-col`;
  const navigationItems = [
    { id: "chat", icon: Sparkles, label: "Chat" },
    { id: "history", icon: HistoryIcon, label: "History" },
    { id: "configuration", icon: SettingsIcon, label: "Configure" },
  ];

  const resetChat = () => {
    setCurrentChatKey(new Date().getTime().toString());
  };

  const resumeChat = (chatKey: string) => {
    setCurrentChatKey(chatKey);
    setActiveView("chat");
  };

  return (
    <div id="kcai-chat" className={cn(!isDetailsPage && 'border-t', containerClass)}>
      <Tabs value={activeView} onValueChange={setActiveView}>
        <div className={cn('flex items-center justify-between px-2 py-2 border-b', isDetailsPage && 'pt-0')}>
          <div className="flex items-center gap-1">
            <TabsList className="h-8">
              {
                navigationItems.map((item) => (
                  <TooltipProvider delayDuration={0}>
                    <Tooltip key={item.id}>
                      <TooltipTrigger asChild>
                        <div>
                          <TabsTrigger value={item.id} >
                            <div className="flex items-center justify-between">
                              <item.icon className="h-4 w-4" />
                              {kcAiChatWindow.width > 800 && <span className='ml-2'>{item.label}</span>}
                            </div>
                          </TabsTrigger>
                        </div>
                      </TooltipTrigger>
                      <TooltipContent side="bottom" hidden={kcAiChatWindow.width > 800}>
                        <p>{item.label}</p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                ))
              }
            </TabsList>
          </div>
          <span className="font-semibold">
            kcAI
            <span className="text-xs align-text-bottom text-gray-500"> (beta)</span>
          </span>
          <div className="flex items-center gap-1">
            <TooltipProvider>
              <Tooltip delayDuration={0}>
                <TooltipTrigger asChild>
                  <Button variant="ghost" size="icon" onClick={resetChat} className="h-8 w-8">
                    <SquarePen className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="bottom" className="px-1.5">
                  New Chat
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>

            {!isFullscreen && onToggleFullscreen && (
              <TooltipProvider>
                <Tooltip delayDuration={0}>
                  <TooltipTrigger asChild>
                    <Button variant="ghost" size="icon" onClick={onToggleFullscreen} className="h-8 w-8">
                      <Maximize2 className="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="bottom" className="px-1.5">
                    Expand
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>

            )}
            {isFullscreen && (
              <>
                {onToggleFullscreen && (
                  <TooltipProvider>
                    <Tooltip delayDuration={0}>
                      <TooltipTrigger asChild>
                        <Button variant="ghost" size="icon" onClick={onToggleFullscreen} className="h-8 w-8">
                          <Minimize2 className="h-4 w-4" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent side="bottom" className="px-1.5">
                        Collapsed
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>


                )}
              </>
            )}
            {onClose && (
              <TooltipProvider>
                <Tooltip delayDuration={0}>
                  <TooltipTrigger asChild>
                    <Button variant="ghost" size="icon" onClick={onClose} className="h-8 w-8">
                      <X className="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="bottom" className="px-1.5">
                    Close chat
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>

            )}
          </div>
        </div>
        <TabsContent value='chat' className={cn(isDetailsPage ? 'chatbot-details-inner-container' : 'chatbot-list-inner-container')}>
          {
            kcAIStoredModelsCollection.providerCollection && Object.keys(kcAIStoredModelsCollection.providerCollection)?.length > 0 ?
              <ChatWindow
                currentChatKey={currentChatKey || ''}
                cluster={cluster}
                config={config}
                isDetailsPage={isDetailsPage}
                namespace={isDetailsPage ? namespace : undefined}
                kcAIStoredModels={kcAIStoredModelsCollection}
                resetChat={resetChat}
              />
              :
              <div className={cn("flex items-center justify-center", isDetailsPage ? 'chatbot-details-inner-container' : 'chatbot-list-inner-container')}>
                <p className="w-3/4 p-4 rounded text-center text-muted-foreground">
                  <span>You haven't set up any providers yet.</span>
                  <br />
                  <span>Click
                    <span className="text-blue-600/100 dark:text-sky-400/100 cursor-pointer" onClick={() => setActiveView('configuration')}> here</span>
                    , to go to Configuration and add one now.</span>
                </p>
              </div>
          }


        </TabsContent>
        <TabsContent value="history" className={cn(isDetailsPage ? 'chatbot-details-inner-container' : 'chatbot-list-inner-container')}>
          <ChatHistory resumeChat={resumeChat} cluster={cluster} config={config} isDetailsPage={isDetailsPage} />
        </TabsContent>
        <TabsContent value="configuration" className={cn(isDetailsPage ? 'chatbot-details-inner-container' : 'chatbot-list-inner-container')}>
          <Configuration cluster={cluster} config={config} setKcAIStoredModelsCollection={setKcAIStoredModelsCollection} isDetailsPage={isDetailsPage} />
        </TabsContent>
      </Tabs>
    </div>
  );

}
