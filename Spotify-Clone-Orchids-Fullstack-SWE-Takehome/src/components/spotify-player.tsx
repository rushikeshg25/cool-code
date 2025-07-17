"use client"

import { useState, useEffect } from 'react'
import { Button } from "@/components/ui/button"
import { Slider } from "@/components/ui/slider"
import { 
  Play, 
  Pause, 
  SkipBack, 
  SkipForward, 
  Shuffle, 
  Repeat, 
  Volume2, 
  VolumeX, 
  Volume1,
  List,
  Monitor,
  Heart,
  PictureInPicture2 
} from 'lucide-react'

interface Track {
  id: string
  title: string
  artist: string
  album: string
  albumArt: string
  duration: number
}

interface SpotifyPlayerProps {
  currentTrack?: Track | null
  isPlaying?: boolean
  currentTime?: number
  volume?: number
  isShuffled?: boolean
  repeatMode?: 'off' | 'all' | 'one'
  onPlayPause?: () => void
  onNext?: () => void
  onPrevious?: () => void
  onShuffle?: () => void
  onRepeat?: () => void
  onSeek?: (time: number) => void
  onVolumeChange?: (volume: number) => void
  onToggleLike?: () => void
  isLiked?: boolean
}

export default function SpotifyPlayer({
  currentTrack = null,
  isPlaying = false,
  currentTime = 0,
  volume = 80,
  isShuffled = false,
  repeatMode = 'off',
  onPlayPause = () => {},
  onNext = () => {},
  onPrevious = () => {},
  onShuffle = () => {},
  onRepeat = () => {},
  onSeek = () => {},
  onVolumeChange = () => {},
  onToggleLike = () => {},
  isLiked = false
}: SpotifyPlayerProps) {
  const [isDragging, setIsDragging] = useState(false)
  const [localCurrentTime, setLocalCurrentTime] = useState(currentTime)
  const [localVolume, setLocalVolume] = useState(volume)
  const [isMuted, setIsMuted] = useState(false)
  const [previousVolume, setPreviousVolume] = useState(volume)

  // Default track when no track is selected
  const defaultTrack: Track = {
    id: 'default',
    title: 'Select a song to play',
    artist: 'Choose from your library',
    album: 'McBeats',
    albumArt: 'https://slelguoygbfzlpylpxfs.supabase.co/storage/v1/object/public/project-uploads/e40f6b6e-b4e9-4096-9ee1-3b7ef5447737/generated_images/small-yellow-burger-icon-with-sesame-see-30215786-20250706215544.jpg?',
    duration: 180
  }

  const activeTrack = currentTrack || defaultTrack

  useEffect(() => {
    if (!isDragging) {
      setLocalCurrentTime(currentTime)
    }
  }, [currentTime, isDragging])

  useEffect(() => {
    setLocalVolume(volume)
  }, [volume])

  const formatTime = (seconds: number) => {
    const minutes = Math.floor(seconds / 60)
    const remainingSeconds = Math.floor(seconds % 60)
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`
  }

  const handleProgressChange = (value: number[]) => {
    const newTime = value[0]
    setLocalCurrentTime(newTime)
    onSeek(newTime)
  }

  const handleVolumeChange = (value: number[]) => {
    const newVolume = value[0]
    setLocalVolume(newVolume)
    onVolumeChange(newVolume)
    if (newVolume > 0 && isMuted) {
      setIsMuted(false)
    }
  }

  const handleMuteToggle = () => {
    if (isMuted) {
      setIsMuted(false)
      setLocalVolume(previousVolume)
      onVolumeChange(previousVolume)
    } else {
      setIsMuted(true)
      setPreviousVolume(localVolume)
      setLocalVolume(0)
      onVolumeChange(0)
    }
  }

  const getVolumeIcon = () => {
    if (isMuted || localVolume === 0) return VolumeX
    if (localVolume < 50) return Volume1
    return Volume2
  }

  const VolumeIcon = getVolumeIcon()

  return (
    <>
      <style jsx>{`
        .custom-slider {
          --slider-track-height: 4px;
          --slider-thumb-size: 12px;
          --slider-track-color: #535353;
          --slider-range-color: #FFC72C;
          --slider-thumb-color: #FFC72C;
          --slider-hover-color: #FFD700;
        }

        .custom-slider .slider-track {
          background-color: var(--slider-track-color);
          height: var(--slider-track-height);
          border-radius: 2px;
          position: relative;
          overflow: hidden;
        }

        .custom-slider .slider-range {
          background-color: var(--slider-range-color);
          height: 100%;
          border-radius: 2px;
        }

        .custom-slider .slider-thumb {
          background-color: var(--slider-thumb-color);
          width: var(--slider-thumb-size);
          height: var(--slider-thumb-size);
          border-radius: 50%;
          border: none;
          box-shadow: none;
          transition: all 0.15s ease;
          opacity: 0;
        }

        .custom-slider:hover .slider-thumb {
          opacity: 1;
        }

        .custom-slider .slider-thumb:hover {
          background-color: var(--slider-hover-color);
          transform: scale(1.2);
        }

        .custom-slider .slider-thumb:active {
          background-color: var(--slider-hover-color);
          transform: scale(1.1);
        }

        .custom-slider:hover .slider-range {
          background-color: var(--slider-hover-color);
        }

        .volume-slider {
          --slider-track-height: 4px;
          --slider-thumb-size: 12px;
        }

        .volume-slider .slider-thumb {
          opacity: 0;
        }

        .volume-slider:hover .slider-thumb {
          opacity: 1;
        }
      `}</style>

      <div className="fixed bottom-0 left-0 right-0 z-50 bg-[#1e1e1e] border-t border-[#2a2a2a] h-[90px] flex items-center px-4">
        <div className="flex items-center justify-between w-full max-w-full">
          
          {/* Left Section - Track Info */}
          <div className="flex items-center space-x-3 min-w-0 flex-1 max-w-[30%]">
            <div className="flex-shrink-0">
              <img 
                src={activeTrack.albumArt} 
                alt={activeTrack.album}
                className="w-14 h-14 rounded-md object-cover"
              />
            </div>
            <div className="min-w-0 flex-1">
              <div className={`text-sm font-medium truncate hover:underline cursor-pointer ${
                currentTrack ? 'text-white' : 'text-[#b3b3b3]'
              }`}>
                {activeTrack.title}
              </div>
              <div className="text-[#b3b3b3] text-xs truncate hover:underline cursor-pointer">
                {activeTrack.artist}
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              className="text-[#b3b3b3] hover:text-white hover:bg-[#2a2a2a] p-2 h-8 w-8"
              onClick={onToggleLike}
              disabled={!currentTrack}
            >
              <Heart className={`h-4 w-4 ${isLiked ? 'fill-[#FFC72C] text-[#FFC72C]' : ''}`} />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="text-[#b3b3b3] hover:text-white hover:bg-[#2a2a2a] p-2 h-8 w-8"
            >
              <PictureInPicture2 className="h-4 w-4" />
            </Button>
          </div>

          {/* Center Section - Playback Controls */}
          <div className="flex flex-col items-center space-y-2 flex-1 max-w-[40%]">
            <div className="flex items-center space-x-2">
              <Button
                variant="ghost"
                size="sm"
                className={`text-[#b3b3b3] hover:text-white hover:bg-[#2a2a2a] p-2 h-8 w-8 ${isShuffled ? 'text-[#FFC72C]' : ''}`}
                onClick={onShuffle}
                disabled={!currentTrack}
              >
                <Shuffle className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="text-[#b3b3b3] hover:text-white hover:bg-[#2a2a2a] p-2 h-8 w-8"
                onClick={onPrevious}
                disabled={!currentTrack}
              >
                <SkipBack className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className={`p-2 h-8 w-8 rounded-full ${
                  currentTrack 
                    ? 'bg-white text-black hover:bg-[#f0f0f0]' 
                    : 'bg-[#535353] text-[#b3b3b3] cursor-not-allowed'
                }`}
                onClick={onPlayPause}
                disabled={!currentTrack}
              >
                {isPlaying ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4 ml-0.5" />}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="text-[#b3b3b3] hover:text-white hover:bg-[#2a2a2a] p-2 h-8 w-8"
                onClick={onNext}
                disabled={!currentTrack}
              >
                <SkipForward className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className={`text-[#b3b3b3] hover:text-white hover:bg-[#2a2a2a] p-2 h-8 w-8 ${repeatMode !== 'off' ? 'text-[#FFC72C]' : ''}`}
                onClick={onRepeat}
                disabled={!currentTrack}
              >
                <Repeat className="h-4 w-4" />
                {repeatMode === 'one' && (
                  <span className="absolute -top-1 -right-1 w-2 h-2 bg-[#FFC72C] rounded-full"></span>
                )}
              </Button>
            </div>
            
            {/* Progress Bar */}
            <div className="flex items-center space-x-3 w-full">
              <span className="text-[#b3b3b3] text-xs font-medium min-w-[35px] text-right">
                {formatTime(localCurrentTime)}
              </span>
              <div className="flex-1 px-2">
                <Slider
                  value={[localCurrentTime]}
                  max={activeTrack.duration}
                  step={1}
                  onValueChange={handleProgressChange}
                  onPointerDown={() => setIsDragging(true)}
                  onPointerUp={() => setIsDragging(false)}
                  disabled={!currentTrack}
                  className="custom-slider w-full"
                />
              </div>
              <span className="text-[#b3b3b3] text-xs font-medium min-w-[35px]">
                {formatTime(activeTrack.duration)}
              </span>
            </div>
          </div>

          {/* Right Section - Volume and Options */}
          <div className="flex items-center space-x-2 flex-1 max-w-[30%] justify-end">
            <Button
              variant="ghost"
              size="sm"
              className="text-[#b3b3b3] hover:text-white hover:bg-[#2a2a2a] p-2 h-8 w-8"
            >
              <List className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="text-[#b3b3b3] hover:text-white hover:bg-[#2a2a2a] p-2 h-8 w-8"
            >
              <Monitor className="h-4 w-4" />
            </Button>
            <div className="flex items-center space-x-2">
              <Button
                variant="ghost"
                size="sm"
                className="text-[#b3b3b3] hover:text-white hover:bg-[#2a2a2a] p-2 h-8 w-8"
                onClick={handleMuteToggle}
              >
                <VolumeIcon className="h-4 w-4" />
              </Button>
              <div className="w-24">
                <Slider
                  value={[localVolume]}
                  max={100}
                  step={1}
                  onValueChange={handleVolumeChange}
                  className="custom-slider volume-slider w-full"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}