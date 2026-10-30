import { useEffect, useState } from 'react'

const BOARD_SIZE = 15
const PLAYER_NAMES = ['あなた', '相手'] as const
type Player = 0 | 1
type Cell = Player | null

function hasFiveInARow(board: Cell[], index: number, player: Player) {
  const row = Math.floor(index / BOARD_SIZE)
  const column = index % BOARD_SIZE
  const directions: [number, number][] = [
    [0, 1],
    [1, 0],
    [1, 1],
    [1, -1],
  ]

  return directions.some(([rowStep, columnStep]) => {
    let count = 1

    for (const sign of [-1, 1]) {
      let nextRow = row + rowStep * sign
      let nextColumn = column + columnStep * sign
      while (
        nextRow >= 0
        && nextRow < BOARD_SIZE
        && nextColumn >= 0
        && nextColumn < BOARD_SIZE
        && board[nextRow * BOARD_SIZE + nextColumn] === player
      ) {
        count += 1
        nextRow += rowStep * sign
        nextColumn += columnStep * sign
      }
    }

    return count >= 5
  })
}

function GameForm() {
  const [board, setBoard] = useState<Cell[]>(() => Array(BOARD_SIZE * BOARD_SIZE).fill(null))
  const [currentPlayer, setCurrentPlayer] = useState<Player | null>(null)
  const [winner, setWinner] = useState<Player | null>(null)
  const [isDraw, setIsDraw] = useState(false)
  const [isSpinning, setIsSpinning] = useState(false)
  const [rouletteLabel, setRouletteLabel] = useState('ルーレットで先手を決めましょう')

  useEffect(() => {
    if (!isSpinning) return

    const intervalId = window.setInterval(() => {
      const randomPlayer = Math.random() < 0.5 ? 0 : 1
      setRouletteLabel(`先手：${PLAYER_NAMES[randomPlayer]}`)
    }, 100)
    const timeoutId = window.setTimeout(() => {
      const selectedPlayer: Player = Math.random() < 0.5 ? 0 : 1
      setCurrentPlayer(selectedPlayer)
      setRouletteLabel(`先手は「${PLAYER_NAMES[selectedPlayer]}」です`)
      setIsSpinning(false)
    }, 1400)

    return () => {
      window.clearInterval(intervalId)
      window.clearTimeout(timeoutId)
    }
  }, [isSpinning])

  function handleCellClick(index: number) {
    if (currentPlayer === null || winner !== null || isDraw || board[index] !== null) return

    const nextBoard = board.slice()
    nextBoard[index] = currentPlayer
    setBoard(nextBoard)

    if (hasFiveInARow(nextBoard, index, currentPlayer)) {
      setWinner(currentPlayer)
    } else if (nextBoard.every((cell) => cell !== null)) {
      setIsDraw(true)
    } else {
      setCurrentPlayer(currentPlayer === 0 ? 1 : 0)
    }
  }

  function handleNewGame() {
    setBoard(Array(BOARD_SIZE * BOARD_SIZE).fill(null))
    setCurrentPlayer(null)
    setWinner(null)
    setIsDraw(false)
    setRouletteLabel('ルーレットで先手を決めましょう')
  }

  const gameFinished = winner !== null || isDraw
  const turnMessage = winner !== null
    ? winner === 0 ? 'あなたの勝ちです！' : '相手の勝ちです！'
    : isDraw
      ? '引き分けです'
      : currentPlayer === null
        ? rouletteLabel
        : currentPlayer === 0
          ? 'あなたの番です'
          : '相手の番です'

  return (
    <main className="game-page">
      <section className="game-screen" aria-labelledby="game-title">
        <header className="game-header">
          <a className="game-back-link" href="/home">ホームへ戻る</a>
          <p className="home-eyebrow">GOMOKU TAISEN</p>
          <h1 id="game-title">五目並べ</h1>
        </header>

        <p className={`game-turn${gameFinished ? ' game-turn-finished' : ''}`} role="status">
          {turnMessage}
        </p>

        <div
          className="gomoku-board"
          role="group"
          aria-label="五目並べの盤面"
        >
          {board.map((cell, index) => {
            const row = Math.floor(index / BOARD_SIZE)
            const column = index % BOARD_SIZE
            const isDisabled = currentPlayer === null || gameFinished || cell !== null

            return (
              <button
                key={index}
                className={`gomoku-cell${cell === null ? '' : ` gomoku-stone-${cell === 0 ? 'black' : 'white'}`}`}
                type="button"
                aria-label={`${row + 1}行${column + 1}列${cell === null ? '空き' : PLAYER_NAMES[cell]}`}
                aria-disabled={isDisabled}
                disabled={isDisabled}
                onClick={() => handleCellClick(index)}
              >
                {cell !== null && <span className="gomoku-stone" aria-hidden="true" />}
              </button>
            )
          })}
        </div>

        <div className="game-controls">
          <button
            className="home-button home-button-primary"
            type="button"
            onClick={() => setIsSpinning(true)}
            disabled={isSpinning || currentPlayer !== null}
          >
            {isSpinning ? 'ルーレット中…' : '先手をルーレットで決める'}
          </button>
          {gameFinished && (
            <button className="home-button home-button-secondary" type="button" onClick={handleNewGame}>
              もう一度遊ぶ
            </button>
          )}
        </div>

        <p className="game-instructions">
          同じ端末で交互に打ちます。黒が「あなた」、白が「相手」です。5つ並べると勝ちです。
        </p>
      </section>
    </main>
  )
}

export default GameForm
