using UnityEngine;

namespace HypeDomino.Game.Inventory
{
    public sealed class CollectionRowSlot : MonoBehaviour
    {
        [SerializeField] CollectionCellSlot[] cells;

        public int Columns => cells.Length;

        public void Configure(CollectionCellSlot[] rowCells) => cells = rowCells;

        public CollectionCellSlot CellAt(int column) => cells[column];
    }
}
