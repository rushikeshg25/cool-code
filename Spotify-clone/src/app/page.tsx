"use client"

import { useState, useEffect, useCallback } from 'react'
import SpotifySidebar from '@/components/spotify-sidebar'
import SpotifyMainContent from '@/components/spotify-main-content'
import SpotifyPlayer from '@/components/spotify-player'
import SpotifyHeader from '@/components/spotify-header'

interface Track {
  id: string
  title: string
  artist: string
  album: string
  albumArt: string
  duration: number
}

export default function SpotifyApp() {
  const [isSidebarVisible, setIsSidebarVisible] = useState(true)
  const [currentView, setCurrentView] = useState<'home' | 'search' | 'library' | 'playlist'>('home')
  const [searchQuery, setSearchQuery] = useState('')
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [volume, setVolume] = useState(80)
  const [isShuffled, setIsShuffled] = useState(false)
  const [repeatMode, setRepeatMode] = useState<'off' | 'all' | 'one'>('off')
  const [isLiked, setIsLiked] = useState(false)
  const [navigationHistory, setNavigationHistory] = useState<string[]>(['home'])
  const [historyIndex, setHistoryIndex] = useState(0)
  const [currentTrack, setCurrentTrack] = useState<Track | null>(null)
  const [recentlyPlayedSongs, setRecentlyPlayedSongs] = useState<Track[]>([]);
  const [madeForYouAlbums, setMadeForYouAlbums] = useState<Track[]>([]);
  const [popularAlbums, setPopularAlbums] = useState<Track[]>([]);

  useEffect(() => {
    const fetchRecentlyPlayed = async () => {
      try {
        const response = await fetch('/api/recently-played');
        if (response.ok) {
          const data = await response.json();
          setRecentlyPlayedSongs(data);
        } else {
          console.error('Failed to fetch recently played songs:', response.statusText);
        }
      } catch (error) {
        console.error('Error fetching recently played songs:', error);
      }
    };

    const fetchMadeForYouAlbums = async () => {
      try {
        const response = await fetch('/api/made-for-you');
        if (response.ok) {
          const data = await response.json();
          setMadeForYouAlbums(data);
        } else {
          console.error('Failed to fetch Made For You albums:', response.statusText);
        }
      } catch (error) {
        console.error('Error fetching Made For You albums:', error);
      }
    };

    const fetchPopularAlbums = async () => {
      try {
        const response = await fetch('/api/popular-albums');
        if (response.ok) {
          const data = await response.json();
          setPopularAlbums(data);
        } else {
          console.error('Failed to fetch Popular albums:', response.statusText);
        }
      } catch (error) {
        console.error('Error fetching Popular albums:', error);
      }
    };

    fetchRecentlyPlayed();
    fetchMadeForYouAlbums();
    fetchPopularAlbums();
  }, []);

  const recordRecentlyPlayed = useCallback(async () => {
    if (currentTrack) {
      try {
        await fetch('/api/recently-played', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            title: currentTrack.title,
            artist: currentTrack.artist,
            album: currentTrack.album,
          }),
        });
      } catch (error) {
        console.error('Failed to record recently played song:', error);
      }
    }
  }, [currentTrack]);

  useEffect(() => {
    if (currentTrack && isPlaying) {
      recordRecentlyPlayed();
    }
  }, [currentTrack, isPlaying, recordRecentlyPlayed]);

  const handlePlayTrack = (track: Track) => {
    setCurrentTrack(track)
    setIsPlaying(true)
    setCurrentTime(0)
  }

  const handleHomeClick = () => {
    setCurrentView('home')
    addToHistory('home')
  }

  const handleSearchClick = () => {
    setCurrentView('search')
    addToHistory('search')
  }

  const handleLibraryToggle = () => {
    setCurrentView('library')
    addToHistory('library')
  }

  const handlePlaylistClick = (playlistId: string) => {
    setCurrentView('playlist')
    addToHistory(`playlist-${playlistId}`)
  }

  const addToHistory = (page: string) => {
    const newHistory = navigationHistory.slice(0, historyIndex + 1)
    newHistory.push(page)
    setNavigationHistory(newHistory)
    setHistoryIndex(newHistory.length - 1)
  }

  const handleBack = () => {
    if (historyIndex > 0) {
      setHistoryIndex(historyIndex - 1)
      const previousPage = navigationHistory[historyIndex - 1]
      if (previousPage === 'home') setCurrentView('home')
      else if (previousPage === 'search') setCurrentView('search')
      else if (previousPage === 'library') setCurrentView('library')
      else if (previousPage.startsWith('playlist-')) setCurrentView('playlist')
    }
  }

  const handleForward = () => {
    if (historyIndex < navigationHistory.length - 1) {
      setHistoryIndex(historyIndex + 1)
      const nextPage = navigationHistory[historyIndex + 1]
      if (nextPage === 'home') setCurrentView('home')
      else if (nextPage === 'search') setCurrentView('search')
      else if (nextPage === 'library') setCurrentView('library')
      else if (nextPage.startsWith('playlist-')) setCurrentView('playlist')
    }
  }

  const handlePlayPause = () => {
    setIsPlaying(!isPlaying)
  }

  const handleNext = () => {
    // Logic for next track
    setCurrentTime(0)
  }

  const handlePrevious = () => {
    // Logic for previous track
    setCurrentTime(0)
  }

  const handleShuffle = () => {
    setIsShuffled(!isShuffled)
  }

  const handleRepeat = () => {
    const modes: ('off' | 'all' | 'one')[] = ['off', 'all', 'one']
    const currentIndex = modes.indexOf(repeatMode)
    const nextIndex = (currentIndex + 1) % modes.length
    setRepeatMode(modes[nextIndex])
  }

  const handleSeek = (time: number) => {
    setCurrentTime(time)
  }

  const handleVolumeChange = (newVolume: number) => {
    setVolume(newVolume)
  }

  const handleToggleLike = () => {
    setIsLiked(!isLiked)
  }

  const handleSearchChange = (query: string) => {
    setSearchQuery(query)
  }

  const handleUserMenuAction = (action: 'profile' | 'settings' | 'logout') => {
    // Handle user menu actions
    console.log('User menu action:', action)
  }

  return (
    <div className="min-h-screen bg-[#121212] text-white overflow-hidden">
      <div className="flex h-screen">
        {/* Sidebar */}
        <SpotifySidebar
          isVisible={isSidebarVisible}
          onToggle={() => setIsSidebarVisible(!isSidebarVisible)}
          onHomeClick={handleHomeClick}
          onSearchClick={handleSearchClick}
          onLibraryToggle={handleLibraryToggle}
          onPlaylistClick={handlePlaylistClick}
          onPlayTrack={handlePlayTrack}
          recentlyPlayedSongs={recentlyPlayedSongs}
        />

        {/* Main Content Area */}
        <div className="flex-1 flex flex-col min-w-0">
          {/* Header */}
          <SpotifyHeader
            showSearch={currentView === 'search'}
            searchQuery={searchQuery}
            onSearchChange={handleSearchChange}
            onBack={handleBack}
            onForward={handleForward}
            canGoBack={historyIndex > 0}
            canGoForward={historyIndex < navigationHistory.length - 1}
            userName="User"
            onUserMenuAction={handleUserMenuAction}
          />

          {/* Main Content */}
          <div className="flex-1 overflow-y-auto pb-24">
            {currentView === 'home' && <SpotifyMainContent onPlayTrack={handlePlayTrack} madeForYouAlbums={madeForYouAlbums} popularAlbums={popularAlbums} />}
            {currentView === 'search' && (
              <div className="p-6">
                <h1 className="text-2xl font-bold mb-6">Search</h1>
                {searchQuery ? (
                  <div>
                    <p className="text-[#b3b3b3]">Search results for "{searchQuery}"</p>
                    {/* Add search results here */}
                  </div>
                ) : (
                  <div>
                    <p className="text-[#b3b3b3]">Search for artists, songs, or podcasts</p>
                    {/* Add browse categories here */}
                  </div>
                )}
              </div>
            )}
            {currentView === 'library' && (
              <div className="p-6">
                <h1 className="text-2xl font-bold mb-6">Your Library</h1>
                <p className="text-[#b3b3b3]">Your playlists and saved music</p>
              </div>
            )}
            {currentView === 'playlist' && (
              <div className="p-6">
                <h1 className="text-2xl font-bold mb-6">Playlist</h1>
                <p className="text-[#b3b3b3]">Selected playlist content</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Bottom Player */}
      <SpotifyPlayer
        currentTrack={currentTrack}
        isPlaying={isPlaying}
        currentTime={currentTime}
        volume={volume}
        isShuffled={isShuffled}
        repeatMode={repeatMode}
        onPlayPause={handlePlayPause}
        onNext={handleNext}
        onPrevious={handlePrevious}
        onShuffle={handleShuffle}
        onRepeat={handleRepeat}
        onSeek={handleSeek}
        onVolumeChange={handleVolumeChange}
        onToggleLike={handleToggleLike}
        isLiked={isLiked}
      />
    </div>
  )
}