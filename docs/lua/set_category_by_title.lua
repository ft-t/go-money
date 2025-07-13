local keywords = { "<Tx1>.COM", "Tx2Name" }

for _, keyword in ipairs(keywords) do
    if string.find(tx:title(), keyword, 1, true) then
        tx:categoryID(4)
        break
    end
end
