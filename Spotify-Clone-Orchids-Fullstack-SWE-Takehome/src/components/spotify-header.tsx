"use client"

import { useState } from 'react'
import { ChevronLeft, ChevronRight, Search, ChevronDown, User, Settings, LogOut } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { 
  DropdownMenu, 
  DropdownMenuContent, 
  DropdownMenuItem, 
  DropdownMenuSeparator, 
  DropdownMenuTrigger 
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar'

interface SpotifyHeaderProps {
  showSearch?: boolean
  searchQuery?: string
  onSearchChange?: (query: string) => void
  onBack?: () => void
  onForward?: () => void
  canGoBack?: boolean
  canGoForward?: boolean
  userAvatar?: string
  userName?: string
  onUserMenuAction?: (action: 'profile' | 'settings' | 'logout') => void
}

export default function SpotifyHeader({
  showSearch = false,
  searchQuery = '',
  onSearchChange,
  onBack,
  onForward,
  canGoBack = false,
  canGoForward = false,
  userAvatar,
  userName = 'User',
  onUserMenuAction
}: SpotifyHeaderProps) {
  const [isDropdownOpen, setIsDropdownOpen] = useState(false)

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onSearchChange?.(e.target.value)
  }

  const handleUserMenuAction = (action: 'profile' | 'settings' | 'logout') => {
    onUserMenuAction?.(action)
    setIsDropdownOpen(false)
  }

  return (
    <header className="sticky top-0 z-50 h-16 bg-background-primary bg-gradient-to-b from-background-secondary to-transparent border-b border-border">
      <div className="h-full flex items-center justify-between px-6">
        {/* Navigation Controls */}
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 rounded-full bg-background-secondary hover:bg-interactive-hover disabled:opacity-50 disabled:cursor-not-allowed"
            onClick={onBack}
            disabled={!canGoBack}
          >
            <ChevronLeft className="h-4 w-4 text-text-primary" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 rounded-full bg-background-secondary hover:bg-interactive-hover disabled:opacity-50 disabled:cursor-not-allowed"
            onClick={onForward}
            disabled={!canGoForward}
          >
            <ChevronRight className="h-4 w-4 text-text-primary" />
          </Button>
        </div>

        {/* Search Bar */}
        <div className="flex-1 max-w-md mx-8">
          {showSearch && (
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-text-secondary" />
              <Input
                type="text"
                placeholder="What do you want to play?"
                value={searchQuery}
                onChange={handleSearchChange}
                className="pl-10 h-10 bg-background-secondary border-border text-text-primary placeholder:text-text-secondary focus:ring-primary focus:border-primary rounded-full"
              />
            </div>
          )}
        </div>

        {/* User Menu */}
        <div className="flex items-center">
          <DropdownMenu open={isDropdownOpen} onOpenChange={setIsDropdownOpen}>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                className="flex items-center gap-2 h-8 px-2 rounded-full bg-background-secondary hover:bg-interactive-hover transition-colors"
              >
                <Avatar className="h-8 w-8">
                  <AvatarImage src={userAvatar} alt={userName} />
                  <AvatarFallback className="bg-primary text-primary-foreground text-sm">
                    <User className="h-4 w-4" />
                  </AvatarFallback>
                </Avatar>
                <ChevronDown className="h-3 w-3 text-text-primary" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent 
              align="end" 
              className="w-48 bg-popover border-border shadow-lg"
            >
              <DropdownMenuItem 
                onClick={() => handleUserMenuAction('profile')}
                className="flex items-center gap-2 text-text-primary hover:bg-interactive-hover cursor-pointer"
              >
                <User className="h-4 w-4" />
                Profile
              </DropdownMenuItem>
              <DropdownMenuItem 
                onClick={() => handleUserMenuAction('settings')}
                className="flex items-center gap-2 text-text-primary hover:bg-interactive-hover cursor-pointer"
              >
                <Settings className="h-4 w-4" />
                Settings
              </DropdownMenuItem>
              <DropdownMenuSeparator className="bg-border" />
              <DropdownMenuItem 
                onClick={() => handleUserMenuAction('logout')}
                className="flex items-center gap-2 text-text-primary hover:bg-interactive-hover cursor-pointer"
              >
                <LogOut className="h-4 w-4" />
                Log out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  )
}