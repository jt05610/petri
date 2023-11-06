import type { FC} from "react";
import { useState } from "react";

type TableRowProps = {
  item: any; // replace with your appropriate type
  handleDelete: () => void;
  handleEdit: (newValues: any) => void; // replace 'any' with your appropriate type
  lastRow: boolean;
}

export const TableRow: FC<TableRowProps> = ({ item, handleDelete, handleEdit, lastRow }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [values, setValues] = useState(item);

  const submitEdit = () => {
    handleEdit(values);
    setIsEditing(false);
  }

  if (isEditing)
    return (
      <tr>
        {/* assuming you're editing 'name' field only for simplicity */}
        <td>
          <input value={values.name}
                 onChange={(e) => setValues({...values, name: e.target.value})}
          />
        </td>
        <td>
          <button onClick={handleDelete}>Delete</button>
          <button onClick={submitEdit}>Submit</button>
        </td>
      </tr>
    );

  return (
    <tr onDoubleClick={() => setIsEditing(true)}>
      <td>{item.name}</td>
      { !lastRow && <td><button onClick={handleDelete}>Delete</button></td> }
    </tr>
  )
}
export default function Table({ initialItems }: { initialItems: any[] }) {
  const [items, setItems] = useState(initialItems);

  const handleDelete = (index: number) =>
    setItems(items => [...items.slice(0, index), ...items.slice(index + 1)]);

  const handleEdit = (index: number, newValues: any) =>
    setItems(items => [...items.slice(0, index), newValues, ...items.slice(index + 1)]);

  return (
    <table>
      <tbody>
      {items.map((item, idx) =>
        (<TableRow
          key={item.id}
          item={item}
          handleDelete={() => handleDelete(idx)}
          handleEdit={(newValues) => handleEdit(idx, newValues)}
          lastRow={idx === items.length - 1}
        />)
      )}
      </tbody>
    </table>
  );
}