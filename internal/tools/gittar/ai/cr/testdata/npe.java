// A Java program to demonstrate that invoking a method
// on null causes NullPointerException
import java.io.*;

class GFG
{
	public static void main (String[] args)
	{
		// Initializing String variable with null value
		String ptr = null;

		// Checking if ptr.equals null or works fine.
		try
		{
			// This line of code throws NullPointerException
			// because ptr is null
			if (ptr.equals("gfg"))
				System.out.print("Same");
			else
				System.out.print("Not Same");
		}
		catch(NullPointerException e)
		{
			System.out.print("NullPointerException Caught");
		}
	}
}
