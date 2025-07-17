"use client"

import { Home, Search, Library, Plus, ArrowRight, Play } from 'lucide-react'
import { useState } from 'react'

interface Track {
  id: string
  title: string
  artist: string
  album: string
  albumArt: string
  duration: number
}

interface PlaylistItem {
  id: string
  title: string
  subtitle: string
  image?: string
  duration?: number
}

interface SpotifySidebarProps {
  isVisible?: boolean
  onToggle?: () => void
  onHomeClick?: () => void
  onSearchClick?: () => void
  onLibraryToggle?: () => void
  onPlaylistClick?: (playlistId: string) => void
  onPlayTrack?: (track: Track) => void
}

export default function SpotifySidebar({
  isVisible = true,
  onToggle,
  onHomeClick,
  onSearchClick,
  onLibraryToggle,
  onPlaylistClick,
  onPlayTrack
}: SpotifySidebarProps) {
  const [isLibraryExpanded, setIsLibraryExpanded] = useState(true)
  const [hoveredPlaylist, setHoveredPlaylist] = useState<string | null>(null)
  
  const recentlyPlayed: PlaylistItem[] = [
    {
      id: '1',
      title: 'Liked Songs',
      subtitle: 'Playlist • 127 songs',
      image: 'https://v3.fal.media/files/panda/kvQ0deOgoUWHP04ajVH3A_output.png',
      duration: 180
    },
    {
      id: '2',
      title: 'Discover Weekly',
      subtitle: 'Made for you',
      image: 'https://v3.fal.media/files/kangaroo/HRayeBi01JIqfkCjjoenp_output.png',
      duration: 210
    },
    {
      id: '3',
      title: 'Daily Mix 1',
      subtitle: 'Indie Rock, Alternative',
      image: 'https://v3.fal.media/files/elephant/N5qDbXOpqAlIcK7kJ4BBp_output.png',
      duration: 225
    },
    {
      id: '4',
      title: 'Chill Hits',
      subtitle: 'Spotify • 147 songs',
      image: 'https://v3.fal.media/files/rabbit/tAQ6AzJJdlEZW-y4eNdxO_output.png',
      duration: 240
    },
    {
      id: '5',
      title: 'My Playlist #1',
      subtitle: 'Playlist • 23 songs',
      image: 'https://v3.fal.media/files/elephant/C_rLsEbIUdbn6nQ0wz14S_output.png',
      duration: 195
    },
    {
      id: '6',
      title: 'Rock Classics',
      subtitle: 'Spotify • 89 songs',
      image: 'https://v3.fal.media/files/rabbit/b11V_uidRMsa2mTr5mCfz_output.png',
      duration: 250
    },
  ]

  const handleLibraryToggle = () => {
    setIsLibraryExpanded(!isLibraryExpanded)
    onLibraryToggle?.()
  }

  const handlePlaylistClick = (playlistId: string) => {
    onPlaylistClick?.(playlistId)
  }

  const handlePlayTrack = (item: PlaylistItem) => {
    const track: Track = {
      id: item.id,
      title: item.title,
      artist: item.subtitle.split(' • ')[0] || 'Playlist',
      album: item.title,
      albumArt: item.image || '/api/placeholder/56/56',
      duration: item.duration || 180
    }
    onPlayTrack?.(track)
  }

  const handlePlayButtonClick = (e: React.MouseEvent, item: PlaylistItem) => {
    e.stopPropagation()
    handlePlayTrack(item)
  }

  if (!isVisible) return null

  return (
    <div className="w-[280px] h-screen bg-[#121212] flex flex-col rounded-lg">
      {/* App Logo */}
      <div className="p-4 border-b border-[#2a2a2a]">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 flex items-center justify-center">
            <img 
              src="https://slelguoygbfzlpylpxfs.supabase.co/storage/v1/object/public/project-uploads/e40f6b6e-b4e9-4096-9ee1-3b7ef5447737/generated_images/small-yellow-burger-icon-with-sesame-see-30215786-20250706215544.jpg?"
              alt="McBeats Burger Logo"
              className="w-8 h-8 object-contain"
            />
          </div>
          <h1 className="text-white font-bold text-xl mcbeats-logo">McBeats</h1>
        </div>
      </div>

      {/* Top Navigation */}
      <div className="p-4 space-y-2">
        <button
          onClick={onHomeClick}
          className="w-full flex items-center gap-3 p-3 rounded-md text-white hover:bg-[#2a2a2a] transition-colors duration-200"
        >
          <Home className="w-5 h-5" />
          <span className="font-medium">Home</span>
        </button>
        
        <button
          onClick={onSearchClick}
          className="w-full flex items-center gap-3 p-3 rounded-md text-white hover:bg-[#2a2a2a] transition-colors duration-200"
        >
          <Search className="w-5 h-5" />
          <span className="font-medium">Search</span>
        </button>
      </div>

      {/* Your Library Section */}
      <div className="px-4 pb-2">
        <button
          onClick={handleLibraryToggle}
          className="w-full flex items-center justify-between p-3 rounded-md text-white hover:bg-[#2a2a2a] transition-colors duration-200"
        >
          <div className="flex items-center gap-3">
            <Library className="w-5 h-5" />
            <span className="font-medium">Your Library</span>
          </div>
          <div className="flex items-center gap-2">
            <Plus className="w-4 h-4" />
            <ArrowRight 
              className={`w-4 h-4 transition-transform duration-200 ${
                isLibraryExpanded ? 'rotate-90' : ''
              }`}
            />
          </div>
        </button>
      </div>

      {/* Playlist Section */}
      {isLibraryExpanded && (
        <div className="flex-1 overflow-y-auto px-4 pb-4">
          <div className="space-y-2">
            <div className="px-3 py-2">
              <h3 className="text-sm font-medium text-[#b3b3b3]">Recently played</h3>
            </div>
            
            {recentlyPlayed.map((item) => (
              <div
                key={item.id}
                className="relative w-full flex items-center gap-3 p-2 rounded-md hover:bg-[#2a2a2a] transition-all duration-200 group border border-transparent hover:border-gray-600/50"
                onMouseEnter={() => setHoveredPlaylist(item.id)}
                onMouseLeave={() => setHoveredPlaylist(null)}
              >
                <button
                  onClick={() => handlePlaylistClick(item.id)}
                  className="flex items-center gap-3 flex-1 min-w-0"
                >
                  <div className="w-12 h-12 bg-[#2a2a2a] rounded-md flex items-center justify-center flex-shrink-0 relative">
                    {item.image ? (
                      <img 
                        src={item.image} 
                        alt={item.title}
                        className="w-full h-full object-cover rounded-md"
                      />
                    ) : (
                      <div className="w-full h-full bg-gradient-to-br from-[#FFC72C] to-[#FFD700] rounded-md flex items-center justify-center">
                        <span className="text-black font-bold text-xs">
                          {item.title.charAt(0)}
                        </span>
                      </div>
                    )}
                  </div>
                  
                  <div className="flex-1 min-w-0 text-left">
                    <p className="text-white font-medium text-sm truncate">
                      {item.title}
                    </p>
                    <p className="text-[#b3b3b3] text-xs truncate">
                      {item.subtitle}
                    </p>
                  </div>
                </button>

                {/* Play button overlay */}
                <div className={`absolute right-2 transition-opacity duration-200 ${
                  hoveredPlaylist === item.id ? 'opacity-100' : 'opacity-0'
                }`}>
                  <button
                    onClick={(e) => handlePlayButtonClick(e, item)}
                    className="w-8 h-8 bg-[#FFC72C] rounded-full flex items-center justify-center shadow-lg transform transition-transform duration-200 hover:scale-110"
                  >
                    <Play className="w-3 h-3 text-black fill-black ml-0.5" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}